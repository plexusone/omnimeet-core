# ADR-008: Frontend SDK Architecture

## Status

Proposed

## Date

2026-07-03

## Context

OmniMeet requires frontend components for human participants to join meetings via web browsers. Two approaches are possible:

1. **Direct provider SDKs** — Use LiveKit React, Daily React, etc. directly
2. **Unified frontend SDK** — Create `@plexusone/omnimeet-js` that abstracts providers

The OmniAgent embedded web UI needs meeting capabilities that work across providers.

## Decision

We will create a unified frontend SDK with provider adapters, following the same pattern as the Go backend.

### Package Structure

```
@plexusone/omnimeet-js        # Core TypeScript library
@plexusone/omnimeet-react     # React components and hooks
@plexusone/omnimeet-livekit   # LiveKit adapter
@plexusone/omnimeet-daily     # Daily adapter
```

### Core TypeScript Library

```typescript
// @plexusone/omnimeet-js

export interface Meeting {
  id: string;
  name: string;
  status: MeetingStatus;
  participants: Participant[];
}

export interface Participant {
  id: string;
  name: string;
  kind: ParticipantKind;
  tracks: Track[];
  isSelf: boolean;
  isSpeaking: boolean;
}

export interface Track {
  id: string;
  kind: TrackKind;
  source: TrackSource;
  muted: boolean;
}

export interface MeetingClient {
  // Connection
  join(url: string, token: string): Promise<void>;
  leave(): Promise<void>;

  // Local participant
  enableMicrophone(): Promise<Track>;
  disableMicrophone(): Promise<void>;
  enableCamera(): Promise<Track>;
  disableCamera(): Promise<void>;
  enableScreenShare(): Promise<Track>;
  disableScreenShare(): Promise<void>;

  // Track control
  muteTrack(trackId: string): Promise<void>;
  unmuteTrack(trackId: string): Promise<void>;

  // Data messages
  sendMessage(data: unknown, reliable?: boolean): Promise<void>;

  // State
  readonly meeting: Meeting | null;
  readonly localParticipant: Participant | null;
  readonly remoteParticipants: Participant[];

  // Events
  on(event: 'participantJoined', handler: (p: Participant) => void): void;
  on(event: 'participantLeft', handler: (p: Participant) => void): void;
  on(event: 'trackPublished', handler: (p: Participant, t: Track) => void): void;
  on(event: 'trackMuted', handler: (p: Participant, t: Track) => void): void;
  on(event: 'activeSpeakerChanged', handler: (p: Participant | null) => void): void;
  on(event: 'messageReceived', handler: (data: unknown, from: Participant) => void): void;
  off(event: string, handler: Function): void;
}

// Factory function
export function createMeetingClient(provider: 'livekit' | 'daily'): MeetingClient;
```

### React Components

```typescript
// @plexusone/omnimeet-react

import { createContext, useContext } from 'react';
import { MeetingClient, Meeting, Participant, Track } from '@plexusone/omnimeet-js';

// Context
export const MeetingContext = createContext<MeetingClient | null>(null);
export function MeetingProvider({ provider, children }: MeetingProviderProps): JSX.Element;

// Hooks
export function useMeeting(): Meeting | null;
export function useLocalParticipant(): Participant | null;
export function useRemoteParticipants(): Participant[];
export function useParticipant(participantId: string): Participant | null;
export function useTracks(participantId: string): Track[];
export function useIsSpeaking(participantId: string): boolean;
export function useConnectionState(): ConnectionState;

// Components
export function MeetingRoom({ url, token, onJoined, onLeft, children }: MeetingRoomProps): JSX.Element;
export function ParticipantTile({ participant, showVideo, showAudio }: ParticipantTileProps): JSX.Element;
export function VideoTrack({ track, mirror }: VideoTrackProps): JSX.Element;
export function AudioTrack({ track }: AudioTrackProps): JSX.Element;
export function Controls({ showMute, showCamera, showScreenShare, showLeave }: ControlsProps): JSX.Element;
export function ParticipantList({ showAgents }: ParticipantListProps): JSX.Element;
export function ActiveSpeaker(): JSX.Element;
```

### Provider Adapters

```typescript
// @plexusone/omnimeet-livekit

import { Room as LiveKitRoom } from 'livekit-client';
import { MeetingClient, Participant, Track } from '@plexusone/omnimeet-js';

export class LiveKitMeetingClient implements MeetingClient {
  private room: LiveKitRoom;

  async join(url: string, token: string): Promise<void> {
    await this.room.connect(url, token);
  }

  async leave(): Promise<void> {
    await this.room.disconnect();
  }

  // Map LiveKit events to OmniMeet events
  private setupEventHandlers() {
    this.room.on('participantConnected', (p) => {
      this.emit('participantJoined', this.mapParticipant(p));
    });
  }

  private mapParticipant(lkParticipant: LKParticipant): Participant {
    return {
      id: lkParticipant.sid,
      name: lkParticipant.name || lkParticipant.identity,
      kind: this.inferKind(lkParticipant),
      tracks: lkParticipant.trackPublications.map(this.mapTrack),
      isSelf: lkParticipant === this.room.localParticipant,
      isSpeaking: lkParticipant.isSpeaking,
    };
  }
}

// Factory registration
import { registerProvider } from '@plexusone/omnimeet-js';
registerProvider('livekit', () => new LiveKitMeetingClient());
```

```typescript
// @plexusone/omnimeet-daily

import DailyIframe from '@daily-co/daily-js';
import { MeetingClient } from '@plexusone/omnimeet-js';

export class DailyMeetingClient implements MeetingClient {
  private daily: DailyIframe.DailyCall;

  async join(url: string, token: string): Promise<void> {
    await this.daily.join({ url, token });
  }

  // ... similar implementation mapping Daily to OmniMeet
}

registerProvider('daily', () => new DailyMeetingClient());
```

### Usage Example

```tsx
// In OmniAgent embedded UI

import { MeetingProvider, MeetingRoom, ParticipantTile, Controls } from '@plexusone/omnimeet-react';
import '@plexusone/omnimeet-livekit'; // Register LiveKit adapter

function AgentMeetingUI({ meetingUrl, token }: Props) {
  return (
    <MeetingProvider provider="livekit">
      <MeetingRoom url={meetingUrl} token={token}>
        <div className="meeting-grid">
          <ParticipantTiles />
        </div>
        <Controls showMute showLeave />
        <TranscriptPanel />
      </MeetingRoom>
    </MeetingProvider>
  );
}

function ParticipantTiles() {
  const participants = useRemoteParticipants();
  const local = useLocalParticipant();

  return (
    <>
      {local && <ParticipantTile participant={local} showVideo showAudio />}
      {participants.map(p => (
        <ParticipantTile key={p.id} participant={p} showVideo showAudio />
      ))}
    </>
  );
}
```

### Build Configuration

```json
// package.json for @plexusone/omnimeet-react
{
  "name": "@plexusone/omnimeet-react",
  "version": "0.1.0",
  "main": "dist/index.js",
  "module": "dist/index.esm.js",
  "types": "dist/index.d.ts",
  "peerDependencies": {
    "@plexusone/omnimeet-js": "^0.1.0",
    "react": "^18.0.0 || ^19.0.0",
    "react-dom": "^18.0.0 || ^19.0.0"
  },
  "devDependencies": {
    "tsup": "^8.0.0",
    "typescript": "^5.0.0"
  }
}
```

## Consequences

### Positive

- **Unified API** — Same components work with LiveKit and Daily
- **OmniAgent integration** — Embedded UI doesn't need provider-specific code
- **Developer experience** — React hooks and components are familiar patterns
- **Consistency** — Matches backend provider pattern

### Negative

- **Abstraction overhead** — Additional layer between app and provider SDK
- **Feature parity** — Must implement features for all providers
- **Maintenance** — Must track changes in provider SDKs

### Mitigations

- Start with minimal feature set (join, audio, video, leave)
- Escape hatch to access underlying provider for advanced features
- Automated tests against provider SDKs

## Alternatives Considered

### Direct Provider SDKs

Let users use LiveKit React, Daily React, etc. directly.

Rejected because:

- OmniAgent UI would need provider-specific code paths
- Switching providers requires significant frontend changes
- Inconsistent with backend architecture

### Wrapper Only (No Components)

Only provide `@plexusone/omnimeet-js`, let users build their own components.

Rejected because:

- Components for participant tiles, controls are standard
- Better DX with ready-to-use components
- Can still use hooks only if preferred

### Web Components

Use framework-agnostic Web Components instead of React.

Considered but deferred because:

- React is already used in OmniAgent UI
- Web Components can be added later for framework diversity
- React has better ecosystem for real-time UI

## Timeline

| Phase | Deliverable | Target |
|-------|-------------|--------|
| V0.5 | `@plexusone/omnimeet-js` core types | After backend V0.3 |
| V0.5 | `@plexusone/omnimeet-livekit` adapter | After backend V0.3 |
| V0.5 | `@plexusone/omnimeet-react` components | After backend V0.3 |
| V0.6 | `@plexusone/omnimeet-daily` adapter | After backend V0.4 |
| V1.0 | Full component library | V1.0 release |

## References

- [LiveKit React Components](https://docs.livekit.io/reference/components/react/)
- [Daily React Hooks](https://docs.daily.co/reference/daily-react)
- [OmniAgent Embedded UI](https://github.com/plexusone/omniagent)
