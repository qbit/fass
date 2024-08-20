package main

import (
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/pawal/go-hass"
)

type Device struct {
	Group string
}

type State struct {
	Devices []Device
}

func getDevice(id string, h *hass.Access) (hass.Device, error) {
	s, err := h.GetState(id)
	if err != nil {
		return nil, err
	}
	return h.GetDevice(s)
}

func makeEntity(e hass.State, h *hass.Access) *widget.Card {
	fmt.Printf("%s: (%s) %s\n", e.EntityID,
		e.Attributes["friendly_name"],
		e.State)

	entityName := fmt.Sprintf("%s", e.Attributes["friendly_name"])
	entityButton := NewToggle()
	if e.State == "on" {
		entityButton.On = true
	}
	entityButton.OnChanged(func(on bool) {
		dev, err := getDevice(e.EntityID, h)
		if err != nil {
			log.Println(err)
			return
		}
		dev.Toggle()
	})

	return widget.NewCard("", entityName, container.NewVBox(
		container.NewHBox(entityButton),
	))
}

func main() {
	token := os.Getenv("HAAS_TOKEN")
	h := hass.NewAccess("https://home.bold.daemon", "")
	h.SetBearerToken(token)
	err := h.CheckAPI()
	if err != nil {
		log.Fatalln(err)
	}

	a := app.New()
	w := a.NewWindow("faas")
	if w == nil {
		log.Fatalln("unable to create window")
	}

	ctrlQ := &desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}
	ctrlW := &desktop.CustomShortcut{KeyName: fyne.KeyW, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlQ, func(shortcut fyne.Shortcut) {
		a.Quit()
	})
	w.Canvas().AddShortcut(ctrlW, func(shortcut fyne.Shortcut) {
		w.Hide()
	})

	lights, err := h.FilterStates("light")
	if err != nil {
		log.Fatalln(err)
	}
	switches, err := h.FilterStates("switch")
	if err != nil {
		log.Fatalln(err)
	}

	var lightCards []fyne.CanvasObject
	var switchCards []fyne.CanvasObject
	for entity := range lights {
		e := lights[entity]
		card := makeEntity(e, h)
		lightCards = append(lightCards, card)
	}

	for entity := range switches {
		e := switches[entity]
		card := makeEntity(e, h)
		switchCards = append(switchCards, card)
	}

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Lights",
			theme.VisibilityIcon(),
			container.NewVBox(
				widget.NewCard("All Lights", "", container.NewVBox(
					widget.NewButton("On", func() {
						for entity := range lights {
							e := lights[entity]
							dev, err := getDevice(e.EntityID, h)
							if err != nil {
								log.Println(err)
								return
							}
							dev.On()
						}
					}),
					widget.NewButton("Off", func() {
						for entity := range lights {
							e := lights[entity]
							dev, err := getDevice(e.EntityID, h)
							if err != nil {
								log.Println(err)
								return
							}
							dev.Off()
						}
					}),
				)),
				container.NewAdaptiveGrid(3, lightCards...),
			)),
		container.NewTabItemWithIcon("Switches",
			theme.RadioButtonIcon(),
			container.NewVBox(
				widget.NewCard("All Switches", "", container.NewVBox(
					widget.NewButton("On", func() {

					}),
					widget.NewButton("Off", func() {
					}),
				)),
				container.NewAdaptiveGrid(3, switchCards...),
			),
		),
	)
	tabs.SetTabLocation(container.TabLocationLeading)
	w.SetContent(tabs)
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	w.ShowAndRun()
}
