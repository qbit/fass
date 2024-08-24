package main

import (
	"fmt"
	"io"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/pawal/go-hass"
)

func loadData(h *hass.Access, lightCards *[]fyne.CanvasObject, switchCards *[]fyne.CanvasObject) {
	lights, err := h.FilterStates("light")
	if err != nil {
		log.Fatalln(err)
	}
	switches, err := h.FilterStates("switch")
	if err != nil {
		log.Fatalln(err)
	}

	for entity := range lights {
		e := lights[entity]
		card := makeEntity(e, h)
		*lightCards = append(*lightCards, card)
	}

	for entity := range switches {
		e := switches[entity]
		card := makeEntity(e, h)
		*switchCards = append(*switchCards, card)
	}
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

func loadSavedData(a fyne.App, input *widget.Entry, file string) {
	uri, err := storage.Child(a.Storage().RootURI(), file)
	if err != nil {
		return
	}

	reader, err := storage.Reader(uri)
	if err != nil {
		return
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return
	}

	input.SetText(string(content))
}

func saveData(a fyne.App, w fyne.Window, input *widget.Entry, file string) {
	uri, err := storage.Child(a.Storage().RootURI(), file)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	writer, err := storage.Writer(uri)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte(input.Text))
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	dialog.ShowInformation("Success", "", w)
}

func main() {
	a := app.New()
	w := a.NewWindow("fass")
	if w == nil {
		log.Fatalln("unable to create window")
	}

	var lightCards []fyne.CanvasObject
	var switchCards []fyne.CanvasObject

	haFile, _ := storage.Child(a.Storage().RootURI(), "haurl")
	haExists, _ := storage.Exists(haFile)
	tokenFile, _ := storage.Child(a.Storage().RootURI(), "hatoken")
	tkExists, _ := storage.Exists(tokenFile)

	urlEntry := widget.NewEntry()
	passEntry := widget.NewPasswordEntry()

	loadSavedData(a, urlEntry, "haurl")
	loadSavedData(a, passEntry, "hatoken")

	h := hass.NewAccess(urlEntry.Text, "")
	if haExists && tkExists {
		h.SetBearerToken(passEntry.Text)
		err := h.CheckAPI()
		if err != nil {
			dialog.ShowError(err, w)
		} else {
			loadData(h, &lightCards, &switchCards)
		}
	}

	settingsForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Home Assistant URL:", Widget: urlEntry},
			{Text: "Access Token:", Widget: passEntry},
			{Text: "", Widget: widget.NewButton("Save", func() {
				saveData(a, w, urlEntry, "haurl")
				saveData(a, w, passEntry, "hatoken")
			})},
		},
	}

	ctrlQ := &desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}
	ctrlW := &desktop.CustomShortcut{KeyName: fyne.KeyW, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlQ, func(shortcut fyne.Shortcut) {
		a.Quit()
	})
	w.Canvas().AddShortcut(ctrlW, func(shortcut fyne.Shortcut) {
		w.Hide()
	})

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Lights",
			theme.VisibilityIcon(),
			container.NewVBox(
				container.NewAdaptiveGrid(3, lightCards...),
			)),
		container.NewTabItemWithIcon("Switches",
			theme.RadioButtonIcon(),
			container.NewVBox(
				container.NewAdaptiveGrid(3, switchCards...),
			),
		),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	w.SetContent(
		container.NewAppTabs(
			container.NewTabItem("Toggles", tabs),
			container.NewTabItem("Settings", container.NewStack(
				settingsForm,
			)),
		),
	)
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	w.ShowAndRun()
}
