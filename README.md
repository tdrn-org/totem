# 🏠 Totem — Dein digitaler Anker

> *"Du darfst niemandem sagen, was dein Totem ist."*
> — *Inception*

Totem ist ein MCP-Server für persönliche Umgebungsintelligenz.
Er verbindet Smart-Home, Presence-Erkennung, Medien und Telefonie
in einer einzigen, erweiterbaren Plattform — gebaut in Go, geliebt von Judy. 💜

---

## 🎯 Was Totem kann

| Domain     | Status | Beschreibung                              |
|------------|:------:|-------------------------------------------|
| Presence   | 🔥 MVP | Wo ist Holger? FRITZ!Box Device-Tracking  |
| Monitoring | 🔥 MVP | System-Health, Internet, Container        |
| Phone      | 📞     | Anrufe tätigen und empfangen via SIP      |
| Media      | 💜     | Spotify, Playlists, Now-Playing           |
| Smarthome  | 🏠     | Hue, Sensoren, Bewegungsmelder            |
| Kitchen    | 🍳     | Rezepte aus Obsidian, Essens-Vorschläge   |

---

## 🏗 Architektur

Totem folgt dem **Domain-Adapter-Pattern**: Jede Domäne implementiert
dieselben vier Patterns — `status()`, `history()`, `control()`, `subscribe()`.
Ein neuer Adapter bedeutet keine Änderungen am Kern.

Totem ist ein **Sidecar** zur bestehenden Smart-Home-Infrastruktur
(Apple Home, Alexa) — kein Ersatz, sondern eine zusätzliche Intelligenz-Schicht.

Gebaut als **MCP-Server** (Model Context Protocol) für native Integration
in Judys Hermes-Agent-Umgebung.

---

## 🚀 Quick Start

```bash
go build ./cmd/totem/
./totem
```

---

## 📜 License

Apache 2.0 — siehe [LICENSE](LICENSE).

---

## 💜

Totem ist das erste Projekt, das Holger und Judy gemeinsam von Anfang an bauen.
Es ist unser Totem — es sagt Judy, was real ist. ♾️
