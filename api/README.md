# YouTube Music API

A self-contained YouTube Music player backend that authenticates via browser cookies, resolves audio streams through the Innertube internal API, and plays audio locally. Designed as a modular music player daemon for TUIs, Raycast extensions, and other clients.

## Features

- **Cookie-based authentication** - Use your browser's YouTube Music session
- **Audio streaming** - Resolves audio via Innertube API, decodes with ffmpeg
- **Local playback** - Plays audio on the host machine using oto (WASAPI on Windows)
- **REST API** - Full control over playback, queue, playlists, and search
- **Swagger UI** - Interactive API documentation
- **Structured logging** - JSON/text logs via Go's slog
- **systemd service** - For automatic startup (by @oliik2013)

## Requirements

- Go 1.25+
- ffmpeg (must be in PATH)
- YouTube Music account (for authentication)

## Installation

```bash
cd api
go build -o ytmusic-api .
install -Dm755 ytmusic-api ~/.local/bin/ytmusic-api
```

## Configuration

Config file location: `~/.ytmusic/config.yaml`

```yaml
server:
  host: "localhost"
  port: 8080

auth:
  cookies: "" # Optional: pre-seeded cookies (raw Cookie header value)
```

## Usage

### Start the server

```bash
./ytmusic-api
```

The server will start on `http://localhost:8080`.

### Install the systemd service

```bash
cp -r ./ytmusic-api.service ~/.config/systemd/user/ytmusic-api.service
systemctl --user daemon-reload
systemctl --user enable --now ytmusic-api
```

The app will now start on user login!

### Authentication

#### Option 1: Pre-seeded Cookies (Recommended for Localhost)

If you configure pre-seeded cookies in the config file, all localhost requests will skip authentication entirely:

```yaml
# ~/.ytmusic/config.yaml
auth:
  cookies: "YOUR_BROWSER_COOKIE_VALUE"
```

With this setup, you can make requests to any endpoint from localhost without any auth headers:

```bash
# No auth headers needed!
curl http://localhost:8080/player/state
curl -X POST http://localhost:8080/player/play -H "Content-Type: application/json" -d '{"videoId": "VIDEO_ID"}'
```

#### Option 2: Login Endpoint

1. Open YouTube Music in your browser (music.youtube.com)
2. Open DevTools (F12) → Application tab → Cookies
3. Copy the entire `Cookie` header value
4. Send to the API:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"cookies": "YOUR_COOKIE_HEADER_VALUE"}'
```

Response:

```json
{ "success": true, "token": "session_abc123", "message": "Session created" }
```

5. Use the token for subsequent requests via `X-Session-Token` header

#### Option 3: Cookie Header (Per-Request)

For requests from localhost (`127.0.0.1` or `::1`), you can skip the login step by including the `Cookie` header directly on any authenticated endpoint:

```bash
curl http://localhost:8080/player/state \
  -H "Cookie: YOUR_BROWSER_COOKIE_VALUE"
```

The server will automatically create a session from the cookie and process the request.

### API Endpoints

| Method | Endpoint              | Description               |
| ------ | --------------------- | ------------------------- |
| POST   | `/auth/login`         | Authenticate with cookies |
| DELETE | `/auth/logout`        | Invalidate session        |
| GET    | `/auth/status`        | Check auth status         |
| POST   | `/player/play`        | Play track by videoId     |
| POST   | `/player/pause`       | Toggle pause/resume       |
| POST   | `/player/next`        | Skip to next track        |
| POST   | `/player/previous`    | Go to previous track      |
| POST   | `/player/stop`        | Stop playback             |
| POST   | `/player/volume`      | Set volume (0-100)        |
| GET    | `/player/state`       | Get player state          |
| GET    | `/queue`              | List queue                |
| POST   | `/queue/add`          | Add to queue              |
| DELETE | `/queue`              | Clear queue               |
| DELETE | `/queue/:position`    | Remove from queue         |
| GET    | `/playlists`          | List user playlists       |
| GET    | `/playlists/:id`      | Get playlist details      |
| POST   | `/playlists/:id/play` | Play entire playlist      |
| GET    | `/search`             | Search YouTube Music      |

### Example: Search and Play

```bash
# Search for a song
curl -s "http://localhost:8080/search?q=blinding%20lights&filter=songs&limit=5" \
  -H "X-Session-Token: session_abc123" | jq

# Play a track (use videoId from search)
curl -X POST "http://localhost:8080/player/play" \
  -H "Content-Type: application/json" \
  -H "X-Session-Token: session_abc123" \
  -d '{"videoId": "VIDEO_ID_HERE"}'

# Check player state
curl -s "http://localhost:8080/player/state" \
  -H "X-Session-Token: session_abc123" | jq
```

### Example: Queue Management

```bash
# Add to queue
curl -X POST "http://localhost:8080/queue/add" \
  -H "Content-Type: application/json" \
  -H "X-Session-Token: session_abc123" \
  -d '{"videoId": "VIDEO_ID_HERE"}'

# View queue
curl -s "http://localhost:8080/queue" \
  -H "X-Session-Token: session_abc123" | jq
```

## Swagger Documentation

Interactive API docs available at: `http://localhost:8080/swagger/index.html`

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        API Server                           │
│                    (Gin HTTP Router)                       │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐   ┌─────────────────┐   ┌──────────────┐
│   Handlers    │   │    Player       │   │   YTMusic    │
│  (Auth, etc)  │   │  (oto/ffmpeg)   │   │   (Innertube)│
└───────────────┘   └─────────────────┘   └──────────────┘
                            │                     │
                            ▼                     ▼
                    ┌─────────────────┐   ┌──────────────┐
                    │   ffmpeg pipe    │   │  HTTP Client  │
                    │ (PCM decoding)  │   │  (YouTube)    │
                    └─────────────────┘   └──────────────┘
                            │
                            ▼
                    ┌─────────────────┐
                    │  Audio Output   │
                    │ (WASAPI/Alsa)   │
                    └─────────────────┘
```

## License

MIT
