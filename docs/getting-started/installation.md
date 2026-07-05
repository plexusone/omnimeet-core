# Installation

## Requirements

- Go 1.26 or later
- A LiveKit server (local or cloud) for development
- Optional: CGO for Opus audio encoding/decoding

## Installing omnimeet-core

Add the core package to your project:

```bash
go get github.com/plexusone/omnimeet-core
```

## Installing a Provider

Install the provider for your platform:

=== "LiveKit"

    ```bash
    go get github.com/plexusone/omni-livekit
    ```

=== "Daily"

    ```bash
    # Coming soon
    go get github.com/plexusone/omni-daily
    ```

## LiveKit Setup

### Option 1: LiveKit Cloud

1. Sign up at [LiveKit Cloud](https://cloud.livekit.io/)
2. Create a new project
3. Copy your API credentials

Set environment variables:

```bash
export LIVEKIT_URL=wss://your-project.livekit.cloud
export LIVEKIT_API_KEY=your-api-key
export LIVEKIT_API_SECRET=your-api-secret
```

### Option 2: Local LiveKit Server

Install and run LiveKit locally:

```bash
# Using Docker
docker run --rm -p 7880:7880 -p 7881:7881 -p 7882:7882/udp \
    livekit/livekit-server \
    --dev \
    --bind 0.0.0.0

# Set environment
export LIVEKIT_URL=ws://localhost:7880
export LIVEKIT_API_KEY=devkey
export LIVEKIT_API_SECRET=secret
```

## CGO Requirements (Optional)

For full audio support with Opus encoding/decoding, you need CGO enabled with the following libraries:

### macOS

```bash
brew install opus libsoxr pkg-config
```

### Ubuntu/Debian

```bash
apt-get install libopus-dev libsoxr-dev pkg-config
```

### Building with CGO

```bash
# Enable CGO and set build tags
CGO_ENABLED=1 go build -tags=cgo,opus ./...
```

### Building without CGO

If CGO is not available, OmniMeet falls back to raw audio passthrough:

```bash
CGO_ENABLED=0 go build ./...
```

!!! note
    Without CGO, audio frames will contain raw data without Opus decoding. This may work for some use cases but is not recommended for production voice applications.

## Verifying Installation

Create a simple test file:

```go
package main

import (
    "fmt"
    "os"

    "github.com/plexusone/omni-livekit/omnimeet"
)

func main() {
    provider, err := omnimeet.NewProvider(omnimeet.Config{
        APIKey:    os.Getenv("LIVEKIT_API_KEY"),
        APISecret: os.Getenv("LIVEKIT_API_SECRET"),
        ServerURL: os.Getenv("LIVEKIT_URL"),
    })
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Provider created: %s\n", provider.Name())
}
```

Run:

```bash
go run main.go
# Output: Provider created: livekit
```

## Next Steps

- [Quick Start](quickstart.md) - Create your first meeting
- [Agent Participation](../guides/agent-participation.md) - Join meetings as an AI agent
