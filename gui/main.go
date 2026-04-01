package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"ytmusic-gui/api"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/yaml.v3"
)

type GUI struct {
	app        fyne.App
	client     *api.Client
	config     *Config
	apiProcess *exec.Cmd

	mainWindow fyne.Window
	nowPlaying *NowPlayingView
	search     *SearchView
	playlists  *PlaylistsView
	queue      *QueueView
	settings   *SettingsView

	currentView string
	contentArea *fyne.Container
	statusLabel *widget.Label
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type AuthConfig struct {
	Cookies string `yaml:"cookies"`
	Token   string `yaml:"token"`
}

type DiscordConfig struct {
	ClientID string `yaml:"client_id"`
}

type AutoStartConfig struct {
	Enabled bool `yaml:"enabled"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Auth      AuthConfig      `yaml:"auth"`
	Discord   DiscordConfig   `yaml:"discord"`
	AutoStart AutoStartConfig `yaml:"autostart"`
}

func (c *Config) GetPort() int {
	if c.Server.Port == 0 {
		return 8080
	}
	return c.Server.Port
}

func (c *Config) GetHost() string {
	if c.Server.Host == "" {
		return "localhost"
	}
	return c.Server.Host
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ytmusic", "config.yaml"), nil
}

func LoadConfig() (*Config, error) {
	cfgPath, err := ConfigPath()
	if err != nil {
		return defaultConfig(), nil
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := defaultConfig()
			os.MkdirAll(filepath.Dir(cfgPath), 0755)
			saveConfig(cfg, cfgPath)
			return cfg, nil
		}
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{}
}

func saveConfig(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (g *GUI) getAPIURL() string {
	return fmt.Sprintf("http://%s:%d", g.config.GetHost(), g.config.GetPort())
}

func (g *GUI) IsAPIOnline() bool {
	resp, err := http.Get(g.getAPIURL() + "/auth/status")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}

func (g *GUI) StartAPI() error {
	if g.apiProcess != nil && g.apiProcess.Process != nil {
		return nil
	}

	cfgPath, _ := ConfigPath()
	if err := saveConfig(g.config, cfgPath); err != nil {
		return err
	}

	apiBinary := filepath.Join(getAPIPath(), "ytmusic-api")
	if _, err := os.Stat(apiBinary); err != nil {
		apiBinary = "ytmusic-api"
	}

	g.apiProcess = exec.Command(apiBinary)
	g.apiProcess.Stdout = os.Stdout
	g.apiProcess.Stderr = os.Stderr
	g.apiProcess.Stdin = os.Stdin

	if err := g.apiProcess.Start(); err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		if g.IsAPIOnline() {
			g.client = api.NewClient(g.getAPIURL(), g.config.Auth.Token)
			if g.config.Auth.Token != "" {
				g.client.Token = g.config.Auth.Token
				g.client.BaseURL = g.getAPIURL()
			}
			return nil
		}
	}

	return fmt.Errorf("API failed to start")
}

func (g *GUI) StopAPI() error {
	if g.apiProcess != nil && g.apiProcess.Process != nil {
		g.apiProcess.Process.Kill()
		g.apiProcess = nil
	}
	g.client = nil
	return nil
}

func (g *GUI) AutoStartEnabled() bool {
	autostartDir := filepath.Join(os.Getenv("HOME"), ".config", "autostart")
	desktopFile := filepath.Join(autostartDir, "ytmusic-gui.desktop")
	_, err := os.Stat(desktopFile)
	return err == nil
}

func (g *GUI) SetAutoStart(enabled bool) error {
	autostartDir := filepath.Join(os.Getenv("HOME"), ".config", "autostart")
	desktopFile := filepath.Join(autostartDir, "ytmusic-gui.desktop")

	if enabled {
		if err := os.MkdirAll(autostartDir, 0755); err != nil {
			return err
		}
		execPath, err := os.Executable()
		if err != nil {
			return err
		}
		content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=YTMusic GUI
Exec=%s
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
`, execPath)
		return os.WriteFile(desktopFile, []byte(content), 0644)
	} else {
		return os.Remove(desktopFile)
	}
}

func getAPIPath() string {
	execPath, _ := os.Executable()
	return filepath.Dir(execPath)
}

func NewGUI() *GUI {
	cfg, err := LoadConfig()
	if err != nil {
		cfg = defaultConfig()
	}

	g := &GUI{
		config:      cfg,
		currentView: "nowplaying",
	}

	g.app = app.New()
	g.app.Settings().SetTheme(theme.DarkTheme())

	win := g.app.NewWindow("YTMusic")
	win.Resize(fyne.NewSize(900, 600))
	win.SetFixedSize(false)
	g.mainWindow = win

	g.statusLabel = widget.NewLabel("Connecting...")
	g.statusLabel.Alignment = fyne.TextAlignCenter

	return g
}

func (g *GUI) buildUI() {
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.MediaPlayIcon(), func() {
			if g.client != nil {
				g.client.PlayPause()
				if g.nowPlaying != nil {
					g.nowPlaying.Refresh()
				}
			}
		}),
		widget.NewToolbarAction(theme.MediaSkipNextIcon(), func() {
			if g.client != nil {
				g.client.Next()
				if g.nowPlaying != nil {
					g.nowPlaying.Refresh()
				}
			}
		}),
		widget.NewToolbarAction(theme.MediaSkipPreviousIcon(), func() {
			if g.client != nil {
				g.client.Previous()
				if g.nowPlaying != nil {
					g.nowPlaying.Refresh()
				}
			}
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ListIcon(), func() {
			g.showView("queue")
		}),
		widget.NewToolbarAction(theme.SearchIcon(), func() {
			g.showView("search")
		}),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			g.showView("playlists")
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			g.showView("settings")
		}),
	)

	g.nowPlaying = NewNowPlayingView(g)
	g.search = NewSearchView(g)
	g.playlists = NewPlaylistsView(g)
	g.queue = NewQueueView(g)
	g.settings = NewSettingsView(g)

	nav := container.New(
		layout.NewVBoxLayout(),
		widget.NewButton("Now Playing", func() { g.showView("nowplaying") }),
		widget.NewButton("Search", func() { g.showView("search") }),
		widget.NewButton("Playlists", func() { g.showView("playlists") }),
		widget.NewButton("Queue", func() { g.showView("queue") }),
		widget.NewButton("Settings", func() { g.showView("settings") }),
	)

	content := container.NewMax()
	g.contentArea = content

	g.showView("nowplaying")

	split := container.NewHSplit(nav, content)
	split.SetOffset(0.15)

	bottomBar := container.NewHBox(
		g.statusLabel,
	)

	mainContent := container.NewVBox(
		toolbar,
		split,
		bottomBar,
	)

	g.mainWindow.SetContent(mainContent)
}

func (g *GUI) showView(view string) {
	g.currentView = view

	var content fyne.CanvasObject
	switch view {
	case "nowplaying":
		content = g.nowPlaying.Build()
	case "search":
		content = g.search.Build()
	case "playlists":
		content = g.playlists.Build()
	case "queue":
		content = g.queue.Build()
	case "settings":
		content = g.settings.Build()
	}

	g.contentArea.Objects = []fyne.CanvasObject{content}
}

func (g *GUI) Run() {
	g.buildUI()

	go func() {
		time.Sleep(500 * time.Millisecond)
		g.updateStatus()
	}()

	g.mainWindow.ShowAndRun()
}

func (g *GUI) updateStatus() {
	if g.IsAPIOnline() {
		g.statusLabel.SetText(fmt.Sprintf("Connected to %s", g.getAPIURL()))
	} else {
		g.statusLabel.SetText("API Offline")
	}
}

func main() {
	gui := NewGUI()
	gui.Run()
}
