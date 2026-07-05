//go:build integration

// Integration tests for MeetingSkill with a real LiveKit provider.
//
// These tests require omni-livekit and should be run from that package
// or with provider injection.
//
// Prerequisites:
//   - Set LIVEKIT_API_KEY, LIVEKIT_API_SECRET, LIVEKIT_URL environment variables
//
// Run from omni-livekit:
//
//	go test -v -tags=integration ./...
package skill

// Integration tests are in omni-livekit/omnimeet/skill_integration_test.go
// since they require the LiveKit provider which cannot be imported here
// (would create circular dependency).
//
// The tests in this file are placeholders that document where the real
// integration tests live.
//
// Test coverage:
// - Unit tests (meeting_test.go): Tool logic with mock provider
// - Integration tests (omni-livekit): Full stack with real LiveKit
// - Smoke tests (omniagent): Wiring validation
