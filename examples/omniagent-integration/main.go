// Example: omniagent-integration demonstrates how to add meeting capabilities to OmniAgent.
//
// This example shows how to:
// 1. Create a meeting provider (LiveKit in this example)
// 2. Create a MeetingSkill with the provider
// 3. Register the skill with OmniAgent
// 4. Allow the agent to create/join meetings via conversation
//
// Prerequisites:
// - LiveKit server running (local or cloud)
// - Set environment variables: LIVEKIT_API_KEY, LIVEKIT_API_SECRET, LIVEKIT_URL
// - OmniAgent configured with LLM provider
//
// Usage:
//
//	go run main.go
//
// Then you can ask the agent things like:
// - "Create a meeting called Team Standup"
// - "Generate a join link for Alice"
// - "List active meetings"
// - "End the meeting"
package main

import (
	"fmt"

	meetingskill "github.com/plexusone/omnimeet-core/skill"
)

func main() {
	fmt.Println("OmniAgent + OmniMeet Integration Example")
	fmt.Println("=========================================")
	fmt.Println()
	fmt.Println("To integrate OmniMeet with OmniAgent, use the following pattern:")
	fmt.Println()
	fmt.Println("```go")
	fmt.Println(`import (
    "github.com/plexusone/omniagent/agent"
    "github.com/plexusone/omni-livekit/omnimeet"
    meetingskill "github.com/plexusone/omnimeet-core/skill"
)

// 1. Create meeting provider
provider, err := omnimeet.NewProvider(omnimeet.Config{
    APIKey:    os.Getenv("LIVEKIT_API_KEY"),
    APISecret: os.Getenv("LIVEKIT_API_SECRET"),
    ServerURL: os.Getenv("LIVEKIT_URL"),
})
if err != nil {
    log.Fatal(err)
}

// 2. Create meeting skill
skill := meetingskill.New(provider, meetingskill.Config{
    DefaultAgentName:   "AI Assistant",
    DefaultMeetingName: "AI Meeting",
    AutoJoinAsAgent:    true, // Agent auto-joins created meetings
})

// 3. Register with OmniAgent
agent, err := agent.New(agentConfig,
    agent.WithCompiledSkill(skill),
)
if err != nil {
    log.Fatal(err)
}

// 4. Agent can now handle meeting requests!`)
	fmt.Println("```")
	fmt.Println()

	// Show available tools
	fmt.Println("Available meeting tools:")
	// Create a nil-provider skill just to list tools
	// (tools are defined statically, don't need provider to list)
	skill := meetingskill.New(nil, meetingskill.Config{})
	for _, tool := range skill.Tools() {
		fmt.Printf("  - %-20s %s\n", tool.Name(), tool.Description())
	}
	fmt.Println()

	fmt.Println("Example conversation:")
	fmt.Println("  User: 'Create a meeting for our team standup'")
	fmt.Println("  Agent: [calls create_meeting tool]")
	fmt.Println("  Agent: 'I've created \"Team Standup\". Join here: https://...'")
	fmt.Println()
	fmt.Println("  User: 'Generate a link for Alice to join'")
	fmt.Println("  Agent: [calls get_join_link tool]")
	fmt.Println("  Agent: 'Here's Alice's join link: https://...'")
	fmt.Println()
	fmt.Println("  User: 'Join the meeting and say hello'")
	fmt.Println("  Agent: [calls join_meeting, then speak_in_meeting]")
	fmt.Println("  Agent: 'I've joined and greeted everyone!'")
}
