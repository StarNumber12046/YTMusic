package main

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"ytmusic-gui/api"
)

type SearchView struct {
	gui      *GUI
	input    *widget.Entry
	list     *widget.List
	results  []api.SearchResult
	selected int
}

func NewSearchView(g *GUI) *SearchView {
	return &SearchView{
		gui: g,
	}
}

func (v *SearchView) Build() fyne.CanvasObject {
	v.input = widget.NewEntry()
	v.input.PlaceHolder = "Search for songs, artists, playlists..."
	v.input.OnSubmitted = func(s string) {
		v.doSearch(s)
	}

	v.list = widget.NewList(
		func() int { return len(v.results) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(v.results) {
				r := v.results[id]
				if r.Track != nil {
					obj.(*widget.Label).SetText(r.Track.Title + " - " + r.Track.Artist)
				} else if r.Playlist != nil {
					obj.(*widget.Label).SetText(r.Playlist.Title)
				}
			}
		},
	)
	v.list.OnSelected = func(id widget.ListItemID) {
		v.selected = id
		if id < len(v.results) {
			r := v.results[id]
			if r.Track != nil && v.gui.client != nil {
				v.gui.client.PlayTrack(r.Track.VideoID)
			} else if r.Playlist != nil && v.gui.client != nil {
				v.gui.client.PlayPlaylist(r.Playlist.ID)
			}
		}
	}

	playNextBtn := widget.NewButton("Play Next", func() {
		if v.gui.client == nil {
			return
		}
		id := v.selected
		if id >= 0 && id < len(v.results) {
			r := v.results[id]
			if r.Track != nil {
				v.gui.client.PlayNext(r.Track.VideoID)
			}
		}
	})

	return container.NewVBox(
		widget.NewLabel("Search (Enter to play)"),
		v.input,
		v.list,
		playNextBtn,
	)
}

func (v *SearchView) doSearch(query string) {
	if v.gui.client == nil {
		return
	}

	results, err := v.gui.client.Search(query)
	if err != nil {
		return
	}

	v.results = results
	v.list.Refresh()
}
