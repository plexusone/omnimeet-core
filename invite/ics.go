package invite

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/plexusone/omnimeet-core/attendee"
)

// ParseICS extracts the meeting from the first event in an iCalendar (.ics)
// file, reading its URL, LOCATION, DESCRIPTION, and conference properties
// such as Google's X-GOOGLE-CONFERENCE.
func ParseICS(r io.Reader) (attendee.JoinTarget, error) {
	lines, err := unfold(r)
	if err != nil {
		return attendee.JoinTarget{}, fmt.Errorf("read ics: %w", err)
	}

	var (
		inEvent bool
		parts   []string
	)
	for _, line := range lines {
		name, value, ok := splitProperty(line)
		if !ok {
			continue
		}
		switch {
		case name == "BEGIN" && strings.EqualFold(value, "VEVENT"):
			inEvent = true
		case name == "END" && strings.EqualFold(value, "VEVENT"):
			if inEvent {
				return Parse(strings.Join(parts, "\n"))
			}
		case inEvent && (name == "URL" || name == "LOCATION" || name == "DESCRIPTION" ||
			(strings.HasPrefix(name, "X-") && strings.Contains(name, "CONFERENCE"))):
			parts = append(parts, unescape(value))
		}
	}
	return attendee.JoinTarget{}, ErrNoMeeting
}

// unfold reads content lines, joining folded lines: a line that starts with a
// space or tab continues the previous one (RFC 5545 section 3.1).
func unfold(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var lines []string
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) && len(lines) > 0 {
			lines[len(lines)-1] += line[1:]
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// splitProperty splits "NAME;PARAM=x:value" into its upper-case name and
// value. A colon inside a quoted parameter value does not end the name.
func splitProperty(line string) (name, value string, ok bool) {
	inQuote := false
	for i, r := range line {
		switch r {
		case '"':
			inQuote = !inQuote
		case ':':
			if !inQuote {
				head := line[:i]
				if semi := strings.IndexByte(head, ';'); semi >= 0 {
					head = head[:semi]
				}
				return strings.ToUpper(head), line[i+1:], true
			}
		}
	}
	return "", "", false
}

var textUnescaper = strings.NewReplacer(`\n`, "\n", `\N`, "\n", `\,`, ",", `\;`, ";", `\\`, `\`)

// unescape decodes iCalendar TEXT escapes.
func unescape(s string) string {
	return textUnescaper.Replace(s)
}
