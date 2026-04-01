package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"ytmusic-gui/api"
)

type QueueView struct {
	gui   *GUI
	list  *widget.List
	queue []api.QueueItem
}

func NewQueueView(g *GUI) *QueueView {
	return &QueueView{
		gui: g,
	}
}

func (v *QueueView) Build() fyne.CanvasObject {
	v.list = widget.NewList(
		func() int { return len(v.queue) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(v.queue) {
				q := v.queue[id]
				obj.(*widget.Label).SetText(q.Track.Title + " - " + q.Track.Artist)
			}
		},
	)
	v.list.OnSelected = func(id widget.ListItemID) {
		if id < len(v.queue) && v.gui.client != nil {
			v.gui.client.PlayTrack(v.queue[id].Track.VideoID)
		}
	}

	clearBtn := widget.NewButton("Clear Queue", func() {
		if v.gui.client != nil {
			v.gui.client.ClearQueue()
			v.loadQueue()
		}
	})

	v.loadQueue()

	return container.NewVBox(
		widget.NewLabel("Queue"),
		clearBtn,
		v.list,
	)
}

func (v *QueueView) loadQueue() {
	if v.gui.client == nil {
		return
	}

	queue, err := v.gui.client.GetQueue()
	if err != nil {
		return
	}

	v.queue = queue
	v.list.Refresh()
}
