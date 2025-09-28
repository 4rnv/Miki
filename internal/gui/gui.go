package main

// WILL BE CHANGED TO PACKAGE GUI AFTER TESTING

import (
	"image/color"
	"miki/internal/yurl"
	"os"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Browser struct {
	Window     fyne.Window
	AddressBar *widget.Entry
	Scroll     *container.Scroll
	LoadBtn    *widget.Button
	Content    *fyne.Container
	Text       string
	Mu         sync.Mutex
}

func NewBrowser(a fyne.App) *Browser {
	win := a.NewWindow("Miki Browser")
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter URL")

	b := &Browser{
		Window:     win,
		AddressBar: urlEntry,
	}
	b.LoadBtn = widget.NewButtonWithIcon("Load", theme.SearchIcon(), func() { go b.LoadAndRender(urlEntry.Text) })
	toolbar := container.New(layout.NewVBoxLayout(), b.AddressBar, b.LoadBtn)
	b.Content = container.NewVBox()
	b.Scroll = container.NewVScroll(b.Content)
	b.Scroll.SetMinSize(fyne.NewSize(800, 600-40))
	contain := container.NewBorder(toolbar, nil, nil, b.Scroll)
	win.SetContent(contain)
	win.Resize(fyne.NewSize(800, 600))
	urlEntry.OnSubmitted = func(s string) {
		go b.LoadAndRender(s)
	}
	return b
}

func (b *Browser) LoadAndRender(raw string) {
	if strings.TrimSpace(raw) == "" {
		return
	}
	u := yurl.NewURL(raw)
	body, err := u.Request(0)
	if err != nil || body == "" {
		b.Special_Page()
		return
	}
	text := yurl.Lex(body)
	b.Mu.Lock()
	b.Text = text
	b.Mu.Unlock()
	content := widget.NewLabel(text)
	content.Wrapping = fyne.TextWrapWord
	b.Content.Objects = []fyne.CanvasObject{content}
}

func (b *Browser) Special_Page() {
	var objs []fyne.CanvasObject
	img := canvas.NewImageFromFile("assets/sayak.png")
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(300, 200))
	objs = append(objs, container.NewCenter(img))
	// ADD BETTER ERROR MESSAGING (PAGE NOT FOUND FOR 404, FORBIDDEN FOR 403 ETC)
	txt := canvas.NewText("This page is either blank or currently unsupported", color.NRGBA{R: 200, G: 20, B: 60, A: 255})
	txt.TextSize = 20
	txt.Alignment = fyne.TextAlignCenter
	objs = append(objs, container.NewCenter(txt))
	b.Mu.Lock()
	b.Content.Objects = objs
	b.Text = ""
	b.Mu.Unlock()
}

// WILL BE CHANGED TO FUNC RUN AFTER TESTING
func main() {
	a := app.NewWithID("miki.browser")
	b := NewBrowser(a)
	if len(os.Args) >= 2 {
		arg := os.Args[1]
		b.AddressBar.SetText(arg)
		go b.LoadAndRender(arg)
	}
	b.Window.ShowAndRun()
}
