package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

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
		if card != nil {
			*lightCards = append(*lightCards, card)
		}
	}

	for entity := range switches {
		e := switches[entity]
		card := makeEntity(e, h)
		if card != nil {
			*switchCards = append(*switchCards, card)
		}
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
	if e.State != "on" && e.State != "off" {
		return nil
	}
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

	card := widget.NewCard("", entityName, container.NewVBox(
		container.NewHBox(entityButton),
	))

	card.Refresh()

	return card
}

func loadSavedData(a fyne.App, w fyne.Window, input *widget.Entry, file string) {
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
		dialog.ShowError(err, w)
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
	certFile, _ := storage.Child(a.Storage().RootURI(), "haCAcert")
	certExists, _ := storage.Exists(certFile)

	urlEntry := widget.NewEntry()
	passEntry := widget.NewPasswordEntry()
	certEntry := widget.NewMultiLineEntry()

	loadSavedData(a, w, urlEntry, "haurl")
	loadSavedData(a, w, passEntry, "hatoken")
	loadSavedData(a, w, certEntry, "haCAcert")

	h := hass.NewAccess(urlEntry.Text, "")
	if haExists && tkExists {
		if certExists && certEntry.Text != "" {
			rootCAs, _ := x509.SystemCertPool()
			if rootCAs == nil {
				rootCAs = x509.NewCertPool()
			}

			if ok := rootCAs.AppendCertsFromPEM([]byte(certEntry.Text)); !ok {
				dialog.ShowError(fmt.Errorf("No certs appended, using system certs only"), w)
			}

			client := &http.Client{
				Timeout: time.Second * 10,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs: rootCAs,
					},
				},
			}
			h.SetClient(client)
		}
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
			{Text: "CA Certificate:", Widget: certEntry},
			{Text: "", Widget: widget.NewButton("Save", func() {
				saveData(a, w, urlEntry, "haurl")
				saveData(a, w, passEntry, "hatoken")
				saveData(a, w, certEntry, "haCAcert")
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

	cols := 5
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Lights",
			theme.VisibilityIcon(),
			container.NewAdaptiveGrid(cols, lightCards...),
		),
		container.NewTabItemWithIcon("Switches",
			theme.RadioButtonIcon(),
			container.NewAdaptiveGrid(cols, switchCards...),
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
