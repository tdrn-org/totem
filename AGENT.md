# AGENT.md — Totem

meta: [[20_Meta/Go-Entwicklung]], [[20_Meta/Projekt-Standards]]

## Vision & Scope

Totem is a smart-home and media hub that unifies devices and services across multiple physical locations. It exposes a single MCP (Model Context Protocol) interface so that AI agents (like Judy) can interact with lights, sensors, media, and presence tracking — regardless of which vendor hardware sits behind each location.

## Technology Stack

- **Language:** Go 1.24+
- **MCP SDK:** `github.com/modelcontextprotocol/go-sdk`
- **HTTP Server:** `github.com/tdrn-org/go-httpserver`
- **Database:** `github.com/tdrn-org/go-database` (SQLite or in-memory)
- **Logging:** `github.com/tdrn-org/go-log`
- **Configuration:** TOML via `github.com/tdrn-org/go-conf`
- **TLS:** `github.com/tdrn-org/go-tlsconf`

## Own Libraries (tdrn-org)

| Library | Purpose |
|---------|---------|
| `go-hue` | Philips Hue Bridge REST API client |
| `go-tr064` | FRITZ!Box TR-064 SOAP service access |
| `go-httpserver` | HTTP server with TLS, CORS, proxy support |
| `go-log` | Structured logging |
| `go-database` | SQLite + schema migrations |
| `go-tlsconf` | TLS configuration |

## Repository Structure

```
totem/
├── AGENT.md
├── ARCHITECTURE.md
├── Makefile
├── version.txt
├── go.mod / go.sum
├── totem.go                  ← CLI entry point (kong)
├── server.go                 ← Server lifecycle, service binding
├── config.go                 ← Logging, HTTP, service wiring
│
├── cmd/totem/main.go         ← main()
│
├── config/                   ← TOML config types
│   ├── config.go
│   ├── server.go
│   ├── store.go
│   ├── location.go
│   ├── logging.go
│   ├── service_fritzbox.go
│   ├── service_huebridge.go
│   └── service_jellyfin.go
│
├── internal/
│   ├── domain/               ← Pure types, NO external imports
│   │   ├── domain.go
│   │   ├── location.go
│   │   ├── user.go
│   │   ├── smarthome/        ← Light, TemperatureSensor, BatteryDevice, MotionSensor
│   │   ├── media/            ← Playlist, MediaItem, MediaService interface
│   │   ├── lan/              ← LanDevice, LanService interface
│   │   └── presence/         ← PresenceResult, PresenceService interface
│   │
│   ├── application/          ← Use-case orchestration
│   │   ├── application.go    ← Host interface
│   │   └── presence.go       ← PresenceTracker (LAN MAC matching + motion)
│   │
│   ├── data/                 ← Persistence layer
│   │   ├── data.go           ← Store
│   │   ├── user.go
│   │   ├── log.go
│   │   └── model/            ← SQLite schema + models
│   │
│   ├── adapters/
│   │   ├── integration/      ← Service interface + Nature constants
│   │   │   └── integration.go
│   │   ├── integration/
│   │   │   ├── fritzbox/     ← FRITZ!Box adapter (LAN + Smarthome)
│   │   │   ├── huebridge/    ← Hue Bridge adapter (Smarthome)
│   │   │   └── jellyfin/     ← Jellyfin adapter (Media)
│   │   └── middleware/
│   │       ├── mcp/          ← MCP tool registration
│   │       │   ├── mcp.go        ← Runtime interface, filterServices[T]
│   │       │   ├── smarthome.go  ← Lights, sensors, motion tools
│   │       │   ├── media.go      ← Playlists, search, recently played
│   │       │   └── presence.go   ← checkUser, checkLocation
│   │       └── rest/         ← REST API (health, docs)
│   │
│   ├── buildinfo/            ← Version, build timestamp
│   └── web/                  ← (reserved for future SvelteKit frontend)
```

## Architecture Principles

### Domain Isolation

`internal/domain/` MUST NOT import anything outside Go stdlib. Verify with:
```bash
go list -json ./internal/domain/... | python3 -c "..."
```

### Nature-Based Service Discovery

Services declare their capabilities via `Natures()` instead of type assertions:

```go
type Nature string
const (
    NatureSmarthome Nature = "smarthome"
    NatureMedia     Nature = "media"
    NatureLan       Nature = "lan"
)
```

The MCP middleware scans services by nature and auto-registers tools. A new service just needs to implement its domain interface and declare its nature — no MCP code changes needed.

### CachedClient + Taint

Network clients (Hue bridge, Jellyfin, FRITZ!Box) are cached behind a mutex. On ANY error, the client is tainted (set to nil) so the next call re-discovers. Zero polling timers needed.

### Location Filtering

All `list*` MCP tools accept an optional `locationID` parameter. Omit to list all locations, specify to filter. Non-breaking — existing callers work unchanged.

## MCP Tools (18 total)

### Smarthome (9 tools)
- `listLights`, `getLight`, `setLight`
- `listTemperatureSensors`, `getTemperatureSensor`
- `listBatteryDevices`, `getBatteryDevice`
- `listMotionSensors`, `getMotion`

### Media (7 tools)
- `listPlaylists`, `getPlaylist`, `createPlaylist`
- `addToPlaylist`, `removeFromPlaylist`
- `searchMedia`, `getRecentlyPlayed`

### Presence (2 tools)
- `checkUser`, `checkLocation`

## Testing

```bash
make check    # gofmt → vet → test → staticcheck
```

Tests exist for `internal/data/`, `internal/data/model/`, `internal/buildinfo/`. MCP tools and adapters are tested via live integration (no mock tests yet).

## Review Checkpoints

When making architectural changes, Judy presents a `[REVIEW REQUIRED]` summary before committing. Holger reviews using H\<n\>-notation.

## Iterative Build Order

| Phase | Status |
|-------|--------|
| Phase 0: Repo skeleton | ✅ Done |
| Phase 1: Domain types + interfaces | ✅ Done |
| Phase 2: MCP tool signatures | ✅ Done |
| Phase 3: SQLite store | ✅ Done |
| Phase 4: Hue Bridge adapter | ✅ Done |
| Phase 5: FRITZ!Box adapter | ✅ Done |
| Phase 6: Jellyfin adapter | ✅ Done |
| Phase 7: Presence tracking | ✅ Done |
| Phase 8: MCP file split + Sensor refactor | ✅ Done (2026-07-04) |
| Phase 9: SvelteKit frontend | 🔮 Future |
