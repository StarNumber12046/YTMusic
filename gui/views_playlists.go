package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"ytmusic-gui/api"
)

type PlaylistsView struct {
	gui       *GUI
	list      *widget.List
	playlists []api.Playlist
	selected  int
}

func NewPlaylistsView(g *GUI) *PlaylistsView {
	return &PlaylistsView{
		gui: g,
	}
}

func (v *PlaylistsView) Build() fyne.CanvasObject {
	v.list = widget.NewList(
		func() int { return len(v.playlists) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(v.playlists) {
				p := v.playlists[id]
				obj.(*widget.Label).SetText(p.Title)
			}
		},
	)
	v.list.OnSelected = func(id widget.ListItemID) {
		v.selected = id
		if id < len(v.playlists) && v.gui.client != nil {
			v.gui.client.PlayPlaylist(v.playlists[id].ID)
		}
	}

	playNextBtn := widget.NewButton("Play Selected Next", func() {
		if v.gui.client == nil {
			return
		}
		id := v.selected
		if id >= 0 && id < len(v.playlists) {
			playlist, err := v.gui.client.GetPlaylist(v.playlists[id].ID)
			if err != nil || len(playlist.Tracks) == 0 {
				return
			}
			v.gui.client.PlayNext(playlist.Tracks[0].VideoID)
		}
	})

	refreshBtn := widget.NewButton("Refresh", func() {
		v.loadPlaylists()
	})

	v.loadPlaylists()

	return container.NewVBox(
		widget.NewLabel("Playlists (select to play)"),
		refreshBtn,
		v.list,
		playNextBtn,
	)
}

func (v *PlaylistsView) loadPlaylists() {
	if v.gui.client == nil {
		return
	}

	playlists, err := v.gui.client.GetPlaylists()
	if err != nil {
		return
	}

	v.playlists = playlists
	v.list.Refresh()
}
