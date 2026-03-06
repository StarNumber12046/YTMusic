package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"ytmusic-tui/api"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF0000")).
			Padding(0, 1)
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

func formatDuration(ms int64) string {
	seconds := ms / 1000
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

type sessionState int

const (
	viewNowPlaying sessionState = iota
	viewSearch
	viewPlaylists
	viewQueue
	viewLyrics
)

type item struct {
	title, desc string
	id          string
	itemType    string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	client      *api.Client
	state       sessionState
	playerState *api.PlayerState
	list        list.Model
	textInput   textinput.Model
	err         error
	width       int
	height      int
	loading     bool
}

type playerStateMsg *api.PlayerState
type errMsg error
type searchResultsMsg []api.SearchResult
type playlistsMsg []api.Playlist
type queueMsg []api.QueueItem
type lyricsMsg *api.LyricsResponse

func (m model) Init() tea.Cmd {
	return tea.Batch(m.tickPlayerState(), textinput.Blink)
}

func (m model) tickPlayerState() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		state, err := m.client.GetPlayerState()
		if err != nil {
			return errMsg(err)
		}
		return playerStateMsg(state)
	})
}

func (m model) searchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := m.client.Search(query)
		if err != nil {
			return errMsg(err)
		}
		return searchResultsMsg(results)
	}
}

func (m model) loadPlaylistsCmd() tea.Cmd {
	return func() tea.Msg {
		playlists, err := m.client.GetPlaylists()
		if err != nil {
			return errMsg(err)
		}
		return playlistsMsg(playlists)
	}
}

func (m model) loadQueueCmd() tea.Cmd {
	return func() tea.Msg {
		queue, err := m.client.GetQueue()
		if err != nil {
			return errMsg(err)
		}
		return queueMsg(queue)
	}
}

func (m model) loadLyricsCmd() tea.Cmd {
	return func() tea.Msg {
		lyrics, err := m.client.GetLyrics()
		if err != nil {
			return errMsg(err)
		}
		return lyricsMsg(lyrics)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-12)

	case playerStateMsg:
		m.playerState = msg
		return m, m.tickPlayerState()

	case searchResultsMsg:
		m.loading = false
		items := make([]list.Item, 0)
		for _, res := range msg {
			if res.Track != nil {
				items = append(items, item{
					title:    res.Track.Title,
					desc:     res.Track.Artist,
					id:       res.Track.VideoID,
					itemType: "track",
				})
			} else if res.Playlist != nil {
				items = append(items, item{
					title:    res.Playlist.Title,
					desc:     fmt.Sprintf("Playlist - %d tracks", res.Playlist.TrackCount),
					id:       res.Playlist.ID,
					itemType: "playlist",
				})
			}
		}
		m.list.SetItems(items)

	case playlistsMsg:
		m.loading = false
		items := make([]list.Item, 0)
		for _, p := range msg {
			items = append(items, item{
				title:    p.Title,
				desc:     fmt.Sprintf("%d tracks", p.TrackCount),
				id:       p.ID,
				itemType: "playlist",
			})
		}
		m.list.SetItems(items)

	case queueMsg:
		m.loading = false
		items := make([]list.Item, 0)
		for _, q := range msg {
			items = append(items, item{
				title:    q.Track.Title,
				desc:     q.Track.Artist,
				id:       q.Track.VideoID,
				itemType: "track",
			})
		}
		m.list.SetItems(items)

	case lyricsMsg:
		m.loading = false
		if msg != nil && msg.ParsedLyrics != nil {
			items := make([]list.Item, 0)
			for _, line := range msg.ParsedLyrics {
				items = append(items, item{
					title:    fmt.Sprintf("[%s] %s", formatDuration(line.TimeMs), line.Text),
					desc:     "",
					id:       "",
					itemType: "lyric",
				})
			}
			m.list.SetItems(items)
		}

	case errMsg:
		m.err = msg
		m.loading = false

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !m.textInput.Focused() {
				return m, tea.Quit
			}
		case "1":
			m.state = viewNowPlaying
		case "2":
			m.state = viewSearch
			m.textInput.Focus()
		case "3":
			m.state = viewPlaylists
			m.loading = true
			return m, m.loadPlaylistsCmd()
		case "4":
			m.state = viewQueue
			m.loading = true
			return m, m.loadQueueCmd()
		case "5":
			m.state = viewLyrics
			m.loading = true
			return m, m.loadLyricsCmd()
		case "enter":
			if m.state == viewSearch && m.textInput.Focused() {
				m.loading = true
				query := m.textInput.Value()
				m.textInput.Blur()
				return m, m.searchCmd(query)
			}
			if i, ok := m.list.SelectedItem().(item); ok {
				if i.itemType == "track" {
					m.client.PlayTrack(i.id)
				} else if i.itemType == "playlist" {
					m.client.PlayPlaylist(i.id)
				}
			}
		case " ":
			m.client.PlayPause()
		case "n":
			m.client.Next()
		case "p":
			m.client.Previous()
		case "s":
			m.client.ToggleShuffle()
		case "r":
			m.client.CycleRepeat()
		case "+", "=":
			if m.playerState != nil {
				m.client.SetVolume(m.playerState.Volume + 5)
			}
		case "-":
			if m.playerState != nil {
				m.client.SetVolume(m.playerState.Volume - 5)
			}
		case "/":
			if m.state == viewSearch {
				m.textInput.Focus()
				return m, nil
			}
		}
	}

	if m.state == viewSearch && m.textInput.Focused() {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(" YOUTUBE MUSIC TUI "))
	s.WriteString("\n\n")

	tabs := []string{"[1] Now Playing", "[2] Search", "[3] Playlists", "[4] Queue", "[5] Lyrics"}
	for i, t := range tabs {
		style := lipgloss.NewStyle().Padding(0, 1)
		if int(m.state) == i {
			style = style.Foreground(lipgloss.Color("#FF0000")).Underline(true)
		}
		s.WriteString(style.Render(t))
	}
	s.WriteString("\n\n")

	switch m.state {
	case viewNowPlaying:
		if m.playerState != nil && m.playerState.CurrentTrack != nil {
			track := m.playerState.CurrentTrack
			s.WriteString(lipgloss.NewStyle().Bold(true).Render("Now Playing:"))
			s.WriteString("\n")
			s.WriteString(fmt.Sprintf("  %s - %s", track.Title, track.Artist))
			s.WriteString("\n\n")

			status := "Paused"
			if m.playerState.IsPlaying && !m.playerState.IsPaused {
				status = "Playing"
			}
			s.WriteString(statusStyle.Render(fmt.Sprintf("[%s]", status)))
			s.WriteString(fmt.Sprintf(" %s / %s", formatDuration(m.playerState.CurrentPositionMs), track.Duration))
			s.WriteString(fmt.Sprintf(" Volume: %d%%", m.playerState.Volume))
			s.WriteString(fmt.Sprintf(" Repeat: %s", m.playerState.Repeat))
			s.WriteString(fmt.Sprintf(" Shuffle: %v", m.playerState.Shuffle))
			s.WriteString("\n\n")
			s.WriteString(" [Space] Play/Pause  [n] Next  [p] Previous  [s] Shuffle  [r] Repeat  [+/-] Volume")
		} else {
			s.WriteString("Nothing playing")
		}
	case viewSearch:
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")
		if m.loading {
			s.WriteString("Searching...")
		} else {
			s.WriteString(m.list.View())
		}
	case viewPlaylists:
		if m.loading {
			s.WriteString("Loading playlists...")
		} else {
			s.WriteString(m.list.View())
		}
	case viewQueue:
		if m.loading {
			s.WriteString("Loading queue...")
		} else {
			s.WriteString(m.list.View())
		}
	case viewLyrics:
		if m.loading {
			s.WriteString("Loading lyrics...")
		} else if len(m.list.Items()) == 0 {
			s.WriteString("No lyrics available for this track")
		} else {
			s.WriteString(m.list.View())
		}
	}

	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return appStyle.Render(s.String())
}

func main() {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Results"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Placeholder = "Search for songs, artists, playlists..."
	ti.CharLimit = 156
	ti.Width = 40

	client := api.NewClient("http://localhost:8080", "")

	m := model{
		client:    client,
		state:     viewNowPlaying,
		list:      l,
		textInput: ti,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
