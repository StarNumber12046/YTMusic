package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type NowPlayingView struct {
	gui    *GUI
	track  *widget.Label
	artist *widget.Label
	status *widget.Label
	volume *widget.ProgressBar
	repeat *widget.Label
}

func NewNowPlayingView(g *GUI) *NowPlayingView {
	return &NowPlayingView{
		gui:    g,
		track:  widget.NewLabel("No track playing"),
		artist: widget.NewLabel(""),
		status: widget.NewLabel("Stopped"),
		volume: widget.NewProgressBar(),
		repeat: widget.NewLabel("Repeat: Off"),
	}
}

func (v *NowPlayingView) Build() fyne.CanvasObject {
	v.refreshData()

	cover := widget.NewCard("Album Art", "", widget.NewLabel("No cover"))

	info := container.NewVBox(
		widget.NewLabel("Now Playing"),
		v.track,
		v.artist,
		v.status,
		v.volume,
		v.repeat,
		widget.NewButton("Play/Pause", func() {
			if v.gui.client != nil {
				v.gui.client.PlayPause()
				v.Refresh()
			}
		}),
		widget.NewButton("Next", func() {
			if v.gui.client != nil {
				v.gui.client.Next()
				v.Refresh()
			}
		}),
		widget.NewButton("Previous", func() {
			if v.gui.client != nil {
				v.gui.client.Previous()
				v.Refresh()
			}
		}),
	)

	return container.NewHBox(
		container.NewVBox(cover),
		container.NewVBox(info),
	)
}

func (v *NowPlayingView) refreshData() {
	if v.gui.client == nil {
		return
	}

	state, err := v.gui.client.GetPlayerState()
	if err != nil {
		return
	}

	if state.CurrentTrack != nil {
		v.track.SetText(state.CurrentTrack.Title)
		v.artist.SetText(state.CurrentTrack.Artist)
	} else {
		v.track.SetText("No track playing")
		v.artist.SetText("")
	}

	if state.IsPlaying {
		v.status.SetText("Playing")
	} else {
		v.status.SetText("Paused")
	}

	v.volume.SetValue(float64(state.Volume))
	v.repeat.SetText("Repeat: " + state.Repeat)
}

func (v *NowPlayingView) Refresh() {
	go func() {
		time.Sleep(100 * time.Millisecond)
		v.refreshData()
	}()
}
