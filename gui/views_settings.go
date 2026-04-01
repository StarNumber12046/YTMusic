package main

import (
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type SettingsView struct {
	gui          *GUI
	portEntry    *widget.Entry
	hostEntry    *widget.Entry
	cookiesEntry *widget.Entry
	discordEntry *widget.Entry
	autostartChk *widget.Check
}

func NewSettingsView(g *GUI) *SettingsView {
	return &SettingsView{
		gui: g,
	}
}

func (v *SettingsView) Build() fyne.CanvasObject {
	cfg := v.gui.config

	v.portEntry = widget.NewEntry()
	v.portEntry.SetText(string(rune(cfg.GetPort() + '0')))
	v.portEntry.PlaceHolder = "Port"

	v.hostEntry = widget.NewEntry()
	v.hostEntry.SetText(cfg.GetHost())
	v.hostEntry.PlaceHolder = "Host"

	v.cookiesEntry = widget.NewEntry()
	v.cookiesEntry.SetText(cfg.Auth.Cookies)
	v.cookiesEntry.PlaceHolder = "Cookies (optional)"
	v.cookiesEntry.Password = true

	v.discordEntry = widget.NewEntry()
	v.discordEntry.SetText(cfg.Discord.ClientID)
	v.discordEntry.PlaceHolder = "Discord Client ID (optional)"

	v.autostartChk = widget.NewCheck("Start with system", func(checked bool) {
		v.gui.SetAutoStart(checked)
	})
	v.autostartChk.Checked = v.gui.AutoStartEnabled()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Host", Widget: v.hostEntry},
			{Text: "Port", Widget: v.portEntry},
			{Text: "Cookies", Widget: v.cookiesEntry},
			{Text: "Discord Client ID", Widget: v.discordEntry},
		},
		OnSubmit: func() {
			v.saveConfig()
		},
	}

	startBtn := widget.NewButton("Start API", func() {
		v.saveConfig()
		v.gui.StartAPI()
	})

	stopBtn := widget.NewButton("Stop API", func() {
		v.gui.StopAPI()
	})

	openSwaggerBtn := widget.NewButton("Open Swagger UI", func() {
		url := v.gui.getAPIURL() + "/swagger/index.html"
		exec.Command("xdg-open", url).Start()
	})

	statusSection := container.NewVBox(
		widget.NewLabel("API Status"),
		container.NewHBox(startBtn, stopBtn, openSwaggerBtn),
	)

	return container.NewVBox(
		widget.NewLabel("Settings"),
		form,
		v.autostartChk,
		widget.NewButton("Save & Apply", func() {
			v.saveConfig()
		}),
		widget.NewSeparator(),
		statusSection,
	)
}

func (v *SettingsView) saveConfig() {
	cfg := v.gui.config

	cfg.Server.Host = v.hostEntry.Text
	cfg.Auth.Cookies = v.cookiesEntry.Text
	cfg.Discord.ClientID = v.discordEntry.Text

	cfgPath, _ := ConfigPath()
	saveConfig(cfg, cfgPath)
}
