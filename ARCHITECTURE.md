# ARCHITECTURE.md — Totem

Architecture Decision Records for Totem. Maintained by Judy, reviewed by Holger.

---

## ADR-001: Nature-Based Service Discovery

**Date:** 2026-06-28
**Status:** Accepted

**Decision:** Services declare their capabilities via a `Natures() []Nature` method instead of relying on Go type assertions in the MCP layer.

**Reason:** A single service can have multiple natures (e.g., FRITZ!Box is both LAN and Smarthome). Type assertions are implicit and panic-prone. Explicit nature declaration makes the capability contract visible and testable.

**Consequences:**
- `integration.Nature` type with constants: `NatureSmarthome`, `NatureMedia`, `NatureLan`
- MCP middleware iterates services by nature → auto-registers tools
- New services only need to implement their domain interface + declare natures
- Generic `filterServices[T]` helper eliminates duplicated filter loops (2026-07-04)

---

## ADR-002: CachedClient + Taint Pattern

**Date:** 2026-06-28
**Status:** Accepted

**Decision:** Network clients (Hue bridge, Jellyfin, FRITZ!Box) are cached behind a mutex. On ANY error from the client, it is tainted (set to nil) so the next call re-discovers.

**Reason:** Re-discovery is cheap (mDNS for Hue, HTTP for Jellyfin) and failures are rare. A polling timer would add complexity without benefit. The taint-on-error pattern is self-healing with zero background goroutines.

**Consequences:**
- `resolveBridgeClient()` / `resolveJellyfinClient()` with mutex-guarded cache
- `taintBridgeClient()` / `taintJellyfinClient()` called on any error
- No polling timers, no health checks, no reconnect logic

---

## ADR-003: Sensor Domain Split

**Date:** 2026-07-04
**Status:** Accepted

**Decision:** Split the monolithic `Sensor` struct into `TemperatureSensor` and `BatteryDevice`.

**Reason:** Temperature measurement and battery level are orthogonal concerns. A motion sensor has a battery but no temperature; a wired thermostat has temperature but no battery. The old `Sensor` used a `-1` sentinel for "no battery" — a code smell.

**Consequences:**
- `TemperatureSensor`: ID, Name, Type, Temperature, Location
- `BatteryDevice`: ID, Name, Type, BatteryLevel (0-100), Location
- `SmarthomeService` interface: `ListTemperatureSensors`, `GetTemperatureSensor`, `ListBatteryDevices`, `GetBatteryDevice`
- MCP tools: `listTemperatureSensors`, `getTemperatureSensor`, `listBatteryDevices`, `getBatteryDevice`
- Old `listSensors`/`getSensor` tools removed (breaking change, but Totem is pre-1.0)

---

## ADR-004: MCP File Split

**Date:** 2026-07-04
**Status:** Accepted

**Decision:** Split the 740-line `mcp.go` monolith into domain-focused files: `mcp.go` (Runtime + handler), `smarthome.go`, `media.go`, `presence.go`.

**Reason:** The single file mixed three domains (smarthome, media, presence) plus formatting functions. Adding a new tool required navigating 700+ lines. Domain-focused files make each domain's tool set immediately visible.

**Consequences:**
- `mcp.go` (84 lines): Runtime interface, `filterServices[T]`, `NewHandler`
- `smarthome.go` (451 lines): 9 tools + formatting
- `media.go` (287 lines): 7 tools + formatting
- `presence.go` (116 lines): 2 tools + formatting

---

## ADR-005: Optional Location Filter on list* Tools

**Date:** 2026-07-04
**Status:** Accepted

**Decision:** All `list*` MCP tools accept an optional `locationID` parameter. Omit to list all locations, specify to filter.

**Reason:** After `checkLocation` returns users at a specific location, the natural next step is "show me the lights at that location." Without a filter, the caller must fetch all lights and filter client-side.

**Consequences:**
- `filterServicesByLocation()` helper in `smarthome.go`
- All four `list*` tools (`listLights`, `listTemperatureSensors`, `listBatteryDevices`, `listMotionSensors`) accept optional `locationID`
- Non-breaking: omitting the parameter returns all locations (backward compatible)

---

## ADR-006: Generic filterServices[T]

**Date:** 2026-07-04
**Status:** Accepted

**Decision:** Replace three duplicated nature-filter loops with a single generic function `filterServices[T any](services, nature, logger) []T`.

**Reason:** `registerSmarthomeTools`, `registerMediaTools`, and `registerPresenceTools` all had identical 15-line loops: iterate services → check nature → type-assert → append. Generics eliminate the duplication without sacrificing type safety.

**Consequences:**
- `filterServices[T]` in `mcp.go`
- Each `register*Tools` function is now 4 lines: filter → check empty → register API
- Type safety preserved: `filterServices[smarthome.SmarthomeService](...)` returns `[]smarthome.SmarthomeService`
