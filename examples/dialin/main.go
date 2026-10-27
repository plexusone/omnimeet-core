// Command dialin joins a meeting by phone from its invitation, plays a short
// tone into the meeting, records what it hears to a WAV file, and leaves.
//
// By default it runs against a simulated meeting from the dialintest package,
// so it needs no credentials:
//
//	go run ./examples/dialin
//	go run ./examples/dialin -invite meeting.ics -out meeting.wav
//
// To join a real meeting, replace the simulated gateway with an omnivoice-core
// voice gateway in audio mode, for example from omni-twilio:
//
//	gw, err := twiliogateway.New(twiliogateway.Config{
//		AccountSID:  os.Getenv("TWILIO_ACCOUNT_SID"),
//		AuthToken:   os.Getenv("TWILIO_AUTH_TOKEN"),
//		PhoneNumber: os.Getenv("TWILIO_PHONE_NUMBER"),
//		PublicURL:   "https://your-server.example.com",
//		Mode:        gateway.PipelineModeAudio,
//	})
//	go gw.Start(ctx) // receives Twilio's webhooks and media stream
//
// and keep the joiner's default timing, which waits for real phone menus to
// play their prompts. The simulated meeting has no prompts, so this example
// uses short pauses.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/plexusone/omnivoice-core/audio/codec"

	"github.com/plexusone/omnimeet-core/attendee"
	"github.com/plexusone/omnimeet-core/dialin"
	"github.com/plexusone/omnimeet-core/dialin/dialintest"
	"github.com/plexusone/omnimeet-core/invite"
)

// sampleInvite is a Zoom invitation for the simulated meeting.
const sampleInvite = `Join Zoom Meeting
https://us02web.zoom.us/j/85012345678?pwd=AbCdEf

Meeting ID: 850 1234 5678
Passcode: 482913

One tap mobile
+16695550100,,85012345678#,,,,*482913# US (San Jose)
`

func main() {
	invitePath := flag.String("invite", "", "invitation text or .ics file (default: a sample Zoom invite)")
	out := flag.String("out", "meeting.wav", "WAV file to record the meeting audio to")
	country := flag.String("country", "US", "country of the dial-in number to call")
	flag.Parse()

	if err := run(*invitePath, *out, *country); err != nil {
		log.Fatal(err)
	}
}

func run(invitePath, out, country string) error {
	target, err := readInvite(invitePath)
	if err != nil {
		return err
	}
	fmt.Printf("Meeting: %s %s\n", target.Platform, target.JoinURL)

	// A simulated meeting that answers the sample invite's dial-in number.
	meeting := &dialintest.Meeting{
		Platform:     attendee.PlatformZoom,
		Number:       "+16695550100",
		ConferenceID: "85012345678",
		Passcode:     "482913",
		Script:       [][]byte{tone(440, 500*time.Millisecond), tone(660, 500*time.Millisecond)},
	}
	gw := dialintest.NewGateway(meeting)

	simulatedMenu := dialin.Timing{Answer: 10 * time.Millisecond, Between: 10 * time.Millisecond}
	joiner := dialin.NewJoiner(gw,
		dialin.WithCountry(country),
		dialin.WithTiming(simulatedMenu),
		dialin.WithKeyDuration(0),
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	sess, err := joiner.Join(ctx, target)
	if err != nil {
		return fmt.Errorf("join: %w", err)
	}

	// Record everything the meeting plays until the call ends.
	type recording struct {
		audio []byte
		err   error
	}
	recorded := make(chan recording, 1)
	go func() {
		audio, err := io.ReadAll(sess.(attendee.AudioReceiver).ReceiveAudio())
		recorded <- recording{audio, err}
	}()

	for ev := range sess.Events() {
		if ev.Type != attendee.EventStateChanged {
			continue
		}
		fmt.Printf("State: %s\n", ev.State)
		if ev.State != attendee.StateAdmitted {
			continue
		}
		// Speak only when the application decides to; here, a short tone.
		if _, err := sess.(attendee.AudioSender).SendAudio().Write(tone(880, 200*time.Millisecond)); err != nil {
			return fmt.Errorf("speak: %w", err)
		}
		// Listen for a moment, then leave.
		time.Sleep(time.Second)
		if err := sess.Leave(ctx); err != nil {
			return fmt.Errorf("leave: %w", err)
		}
	}

	rec := <-recorded
	if rec.err != nil {
		return fmt.Errorf("record: %w", rec.err)
	}
	if err := os.WriteFile(out, codec.MulawToWAV(rec.audio), 0o600); err != nil {
		return fmt.Errorf("write recording: %w", err)
	}
	fmt.Printf("Recorded %.1f seconds of meeting audio to %s\n", float64(len(rec.audio))/8000, out)
	if calls := meeting.Calls(); len(calls) > 0 {
		fmt.Printf("Keys entered: %s\n", calls[0].Keys())
	}
	return nil
}

// readInvite parses the invitation at path, or the sample invite when path
// is empty.
func readInvite(path string) (attendee.JoinTarget, error) {
	if path == "" {
		return invite.Parse(sampleInvite)
	}
	f, err := os.Open(path) //nolint:gosec // the invite path is chosen by the user
	if err != nil {
		return attendee.JoinTarget{}, fmt.Errorf("open invite: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("close invite: %v", err)
		}
	}()

	if strings.HasSuffix(strings.ToLower(path), ".ics") {
		return invite.ParseICS(f)
	}
	text, err := io.ReadAll(f)
	if err != nil {
		return attendee.JoinTarget{}, fmt.Errorf("read invite: %w", err)
	}
	return invite.Parse(string(text))
}

// tone returns a sine tone as mu-law 8 kHz audio.
func tone(hz float64, d time.Duration) []byte {
	pcm := make([]int16, int(d.Seconds()*8000))
	for i := range pcm {
		pcm[i] = int16(0.3 * math.MaxInt16 * math.Sin(2*math.Pi*hz*float64(i)/8000))
	}
	return codec.MulawEncode(pcm)
}
