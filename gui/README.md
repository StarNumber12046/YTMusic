# YTMusic GUI

A full-featured desktop GUI application for YouTube Music with playback controls, search, playlist management, and more.

## Features

- **Now Playing** - View current track with playback controls (play/pause, next, previous)
- **Search** - Search for songs, artists, and playlists
- **Playlists** - Browse and play your YouTube Music playlists
- **Queue** - View and manage the current playback queue
- **Settings**
  - Start/stop the API server in the background
  - Configure API settings (host, port)
  - Set authentication cookies
  - Configure Discord RPC client ID
  - Auto-start with system
  - Open Swagger UI in browser

## Requirements

- Go 1.25+
- GTK3 development libraries (for Fyne)

### Ubuntu/Debian
```bash
sudo apt-get install libgtk-3-dev libgl1-mesa-glx
```

### Fedora/RHEL
```bash
sudo dnf install gtk3-devel mesa-libGLU
```

### macOS
```bash
brew install gtk+3
```

## Build

```bash
cd gui
go build -o ytmusic-gui .
```

## Run

```bash
./ytmusic-gui
```

Or run directly with:
```bash
go run .
```

## Configuration

The GUI saves configuration to `~/.ytmusic/config.yaml`.

## Auto-Start

Enable "Start with system" in Settings to automatically launch the GUI on system startup.

## Note

The API binary (`ytmusic-api`) needs to be in the same directory as the GUI, or in your PATH.
