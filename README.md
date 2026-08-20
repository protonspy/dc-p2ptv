# dc-p2ptv

A federated screen-sharing and video-call network. Anyone can run a node, host their own
community on it, and have the rest of the network carry the bandwidth.

Think of Discord's Go Live, except the server is yours and the upload cost is shared.

## The problem

A central relay pays for every viewer. From the README of this project's predecessor,
`discord-screen-p2p`: at 2.5 Mb/s, five viewers already cost 12.5 Mb/s of upload; at
8 Mb/s, forty. The bill grows with the audience and lands on one machine.

Two things bring it down:

- **Peers talk directly** while a room is small — up to four participants, no media
  server in the path at all.
- **Nodes forward in a tree** once it is not. The origin sends one copy per neighbouring
  node instead of one per viewer.

## How it works

A room holds up to eight participants — camera, microphone, and anyone may share a
screen — plus dozens of viewers who only receive. A viewer can be promoted to
participant, which is the same thing as joining the voice channel.

```
  small room                          larger room
  ┌────────┐   media    ┌────────┐    ┌────────┐        ┌──── node ──── viewers
  │ peer A │◄──────────►│ peer B │    │ peer A │───────►│ node │
  └───┬────┘            └────┬───┘    └────────┘        └──┬───┘
      │  signalling only     │                            │
      └────► control ◄───────┘                     one copy per node,
             plane                                 not per viewer
```

Every node runs the same binary. What differs is the role it plays for a given room:
the **owner node** hosts a community, creates its rooms and holds their key; a **relay
node** forwards media for rooms belonging to somebody else, and cannot open them.

## Trust model

Read this before running a node, and before joining a room on somebody else's.

- A **relay node cannot decrypt** the media it carries. Frames are sealed with a group
  key that never leaves the owning community.
- A **relay node decides nothing** — not who joins, not which room, not with which key.
  It verifies a signed ticket and forwards. Authority lives in the control plane.
- The **owner of a node can watch** the rooms of the community they host. They are that
  community's administrator, the same position as whoever runs a Matrix homeserver. The
  interface says so where people join.
- Any node in the path **sees metadata**: who is in a room, when, and from which IP
  address. Encryption does not hide that, and nothing can.

The reasoning behind each of these is in [`docs/adr/`](docs/adr/).

## Status

Early. The plan and the decisions are written; the protocol is not implemented yet.

What exists today is the repository skeleton, the vocabulary, the adopted stack and four
architecture decision records. Progress is tracked in
[`plans/sala-tempo-real.md`](plans/sala-tempo-real.md).

## Layout

```
server/     Go — control plane, forwarding node, routing, tree
client/     TypeScript — room, mesh, capture, frame sealing
specs/      one feature per directory: requirements, design, tasks
plans/      the decomposition above the specs
docs/       glossary, stack, decision records, and the research wiki
```

## Getting started

```bash
# server
cd server && go build ./... && go test ./...

# client
cd client && npm install && npm test
```

Linters, formatters and the per-package test commands are listed in
[`.claude/rules/project.md`](.claude/rules/project.md).

## Documentation

- [`plans/sala-tempo-real.md`](plans/sala-tempo-real.md) — what is being built, and in
  what order
- [`docs/adr/`](docs/adr/) — the decisions that are expensive to reverse, and why
- [`docs/glossary.md`](docs/glossary.md) — one canonical term per concept
- [`docs/stack.md`](docs/stack.md) — every adopted technology, with the line that earned
  it its place
- [`docs/wiki/`](docs/wiki/) — the research this design rests on: P2P video distribution,
  WebRTC, SFU topologies, and how Discord actually does it

Documentation under `specs/`, `plans/` and `docs/` is written in Portuguese. Code,
identifiers and this README are in English.

## License

See [LICENSE](LICENSE).
