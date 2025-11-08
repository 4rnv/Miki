package gui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"miki/internal/assets"
	"miki/internal/lexer"
	"miki/internal/mtheme"
	"miki/internal/yurl"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Browser struct {
	Window      fyne.Window
	AddressBar  *widget.Entry
	Scroll      *container.Scroll
	LoadBtn     *widget.Button
	Content     *fyne.Container
	Title       *widget.Label
	Text        string
	Mu          sync.Mutex
	History     []string
	HistoryFile string
	CurrentURL  yurl.URL
}

func NewBrowser(a fyne.App) *Browser {
	win := a.NewWindow("Miki Browser")
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter URL")
	title := widget.NewLabel("New Window")
	title.Truncation = fyne.TextTruncateEllipsis
	const (
		InitialWidth  = 800
		InitialHeight = 600
	)
	historyPath := filepath.Join(fyne.CurrentApp().Storage().RootURI().Path(), "history.json")
	fmt.Println("History path: ", historyPath)
	b := &Browser{
		Window:      win,
		AddressBar:  urlEntry,
		Title:       title,
		HistoryFile: historyPath,
	}
	b.loadHistory()
	b.LoadBtn = widget.NewButtonWithIcon("Load", theme.SearchIcon(), func() { go b.LoadAndRender(urlEntry.Text) })
	historyBtn := widget.NewButtonWithIcon("History", theme.HistoryIcon(), func() { b.showHistory() })
	currentFontSize := a.Preferences().FloatWithFallback("font_size", 14)
	currentTheme := a.Preferences().StringWithFallback("theme", "custom")
	customTheme := &mtheme.MTheme{FontSize: float32(currentFontSize)}

	switch currentTheme {
	case "light":
		a.Settings().SetTheme(&mtheme.TemeVariant{Theme: theme.DefaultTheme(), Variant: theme.VariantLight})
	case "dark":
		a.Settings().SetTheme(&mtheme.TemeVariant{Theme: theme.DefaultTheme(), Variant: theme.VariantDark})
	default:
		a.Settings().SetTheme(customTheme)
	}

	showSettingsButton := func() {
		themeSelect := widget.NewSelect([]string{"light", "dark", "custom"}, func(selected string) {
			switch selected {
			case "light":
				a.Settings().SetTheme(&mtheme.TemeVariant{Theme: theme.DefaultTheme(), Variant: theme.VariantLight})
				a.Preferences().SetString("theme", "light")
			case "dark":
				a.Settings().SetTheme(&mtheme.TemeVariant{Theme: theme.DefaultTheme(), Variant: theme.VariantDark})
				a.Preferences().SetString("theme", "dark")
			default:
				a.Settings().SetTheme(customTheme)
				a.Preferences().SetString("theme", "custom")
			}
		})
		themeSelect.SetSelected(currentTheme)
		content := container.NewGridWithColumns(2, widget.NewLabel("Theme"), themeSelect)
		dialog.ShowCustom("Settings", "Close",
			content,
			b.Window,
		)
	}
	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), showSettingsButton)
	toolbar := container.New(newToolbarLayout(),
		container.NewStack(title),
		urlEntry,
		container.NewHBox(b.LoadBtn, historyBtn, settingsBtn),
	)
	b.Content = container.NewVBox()
	b.Scroll = container.NewVScroll(b.Content)
	b.Scroll.SetMinSize(fyne.NewSize(InitialWidth, InitialHeight-40))
	contain := container.NewBorder(toolbar, nil, nil, nil, b.Scroll)
	win.SetContent(contain)
	win.Resize(fyne.NewSize(InitialWidth, InitialHeight))
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
		b.Special_Page(err)
		return
	}
	b.Mu.Lock()
	b.CurrentURL = u
	b.Text = ""
	fyne.DoAndWait(func() {
		b.Title.SetText("Loading...")
		b.addToHistory(u.Raw)
	})
	b.Mu.Unlock()
	if u.Scheme == "view-source" || u.Scheme == "data" || u.Scheme == "file" {
		b.RenderStringContent(body, u.Scheme)
		return
	}
	textOrHTML := lexer.LexTokens(body)
	b.renderTokens(textOrHTML)
}

func (b *Browser) loadAndRenderImage(src string, placeholder *fyne.Container) {
	baseURL := b.CurrentURL
	imgURL := baseURL.ResolveURL(src)
	imageData, err := imgURL.FetchImage()
	if err != nil {
		fmt.Println("Failed to load image:", err)
		b.Mu.Lock()
		placeholder.Objects = []fyne.CanvasObject{
			widget.NewLabel("[Image failed to load]"),
		}
		fyne.DoAndWait(func() {
			placeholder.Refresh()
		})
		b.Mu.Unlock()
		return
	}
	fmt.Println("This function loadAndRenderImage has been triggered for ", imgURL.Raw)
	img := canvas.NewImageFromReader(bytes.NewReader(imageData), src)
	img.FillMode = canvas.ImageFillOriginal
	img.SetMinSize(fyne.NewSize(200, 200))
	b.Mu.Lock()
	placeholder.Objects = []fyne.CanvasObject{img}
	fyne.DoAndWait(func() {
		placeholder.Refresh()
	})
	b.Mu.Unlock()
}

func (b *Browser) Special_Page(err error) {
	var objs []fyne.CanvasObject
	img_res := assets.Get("sayak.png")
	img := canvas.NewImageFromResource(img_res)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(300, 200))
	objs = append(objs, container.NewCenter(img))
	var error_txt string
	if err == nil {
		error_txt = "This page is either blank or currently unsupported"
	} else {
		error_txt = err.Error()
	}
	error_message := widget.NewRichText(&widget.TextSegment{
		Text: error_txt,
		Style: widget.RichTextStyle{
			Alignment: fyne.TextAlignCenter,
			ColorName: theme.ColorNameError,
			SizeName:  fyne.ThemeSizeName("subHeadingText"),
		},
	})
	error_message.Wrapping = fyne.TextWrapWord
	objs = append(objs, error_message)
	b.Mu.Lock()
	b.Content.Objects = objs
	b.Text = ""
	b.Mu.Unlock()
}

func (b *Browser) renderTokens(tokens []lexer.Token) {
	blocks := []fyne.CanvasObject{}
	inline := newRichInline()

	pushBlock := func(rt *widget.RichText) {
		rt.Wrapping = fyne.TextWrapWord
		blocks = append(blocks, rt)
	}

	flushInline := func() {
		if inline != nil && len(inline.Segments) > 0 {
			pushBlock(inline)
			inline = newRichInline()
		}
	}

	type styleState struct{ bold, italic, link, monospace, inTitle, inBody bool }
	state := styleState{}
	stack := []string{}

	beginInline := func() {
		if inline == nil {
			inline = newRichInline()
		}
	}

	for _, tok := range tokens {
		switch tok.Kind {
		case lexer.TextTok:
			if state.inTitle {
				fyne.DoAndWait(func() {
					b.Title.SetText(strings.TrimSpace(tok.Data))
				})
				continue
			}
			if !state.inBody {
				continue
			}
			beginInline()
			seg := &widget.TextSegment{Text: tok.Data}
			seg.Style = widget.RichTextStyleInline
			seg.Style.TextStyle = fyne.TextStyle{
				Bold:      state.bold,
				Italic:    state.italic,
				Monospace: state.monospace,
			}
			if state.link {
				dummy_URL, _ := url.Parse("#") //Will add href later
				hyper_seg := &widget.HyperlinkSegment{
					Text: seg.Text,
					URL:  dummy_URL,
				}
				inline.Segments = append(inline.Segments, hyper_seg)
			} else {
				inline.Segments = append(inline.Segments, seg)
			}
		case lexer.StartTagTok:
			switch tok.Data {
			case "body":
				state.inBody = true
				stack = append(stack, tok.Data)
			case "title":
				state.inTitle = true
				stack = append(stack, tok.Data)
			case "b", "strong":
				state.bold = true
				stack = append(stack, tok.Data)
			case "i", "em":
				state.italic = true
				stack = append(stack, tok.Data)
			case "code":
				state.monospace = true
				stack = append(stack, tok.Data)
			case "p":
				flushInline()
			case "h1":
				flushInline()
				stack = append(stack, "h1")
				inline = widget.NewRichText()
			case "a":
				state.link = true
				stack = append(stack, tok.Data)
			case "pre":
				flushInline()
				stack = append(stack, "pre")
				inline = widget.NewRichText()
				inline.Wrapping = fyne.TextWrapOff
			case "img":
				if src, ok := tok.Attrs["src"]; ok {
					flushInline()
					loadingLabel := widget.NewLabel("Loading image...")
					loadingLabel.TextStyle = fyne.TextStyle{TabWidth: 10, Symbol: true}
					placeholder := container.NewStack(loadingLabel)
					blocks = append(blocks, placeholder)

					go b.loadAndRenderImage(src, placeholder)
				}
				stack = append(stack, tok.Data)

			default:
				stack = append(stack, tok.Data)
			}

		case lexer.EndTagTok:
			switch tok.Data {
			case "body":
				state.inBody = false
				pop(&stack, tok.Data)
			case "title":
				state.inTitle = false
				pop(&stack, tok.Data)
			case "b", "strong":
				state.bold = false
				pop(&stack, tok.Data)
			case "i", "em":
				state.italic = false
				pop(&stack, tok.Data)
			case "code":
				state.monospace = false
				pop(&stack, tok.Data)
			case "p":
				flushInline()
				pop(&stack, tok.Data)
			case "h1":
				if inline != nil {
					txt := collectText(inline)
					heading := widget.NewRichText(&widget.TextSegment{
						Text: txt,
						Style: widget.RichTextStyle{
							SizeName:  fyne.ThemeSizeName("headingText"),
							Alignment: fyne.TextAlignCenter,
							TextStyle: fyne.TextStyle{Bold: true},
						},
					})
					heading.Wrapping = fyne.TextWrapWord
					blocks = append(blocks, heading)
				}
				inline = newRichInline()
				pop(&stack, tok.Data)
			case "a":
				state.link = false
				pop(&stack, tok.Data)
			case "pre":
				if inline != nil {
					rt := inline
					mono := &widget.TextSegment{
						Text:  collectText(rt),
						Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Monospace: true}},
					}
					rt.Segments = []widget.RichTextSegment{mono}
					rt.Wrapping = fyne.TextWrapOff
					blocks = append(blocks, rt)
				}
				inline = newRichInline()
				pop(&stack, tok.Data)
			default:
				pop(&stack, tok.Data)
			}
		}
	}
	flushInline()

	b.Mu.Lock()
	b.Content.Objects = blocks
	b.Mu.Unlock()
}

func newRichInline() *widget.RichText {
	rt := widget.NewRichText()
	rt.Wrapping = fyne.TextWrapWord
	return rt
}

func collectText(rt *widget.RichText) string {
	var sb strings.Builder
	for _, s := range rt.Segments {
		if ts, ok := s.(*widget.TextSegment); ok {
			sb.WriteString(ts.Text)
		}
	}
	return sb.String()
}

// Special func for view-source, data, file schemes
func (b *Browser) RenderStringContent(text string, scheme string) {
	mono := widget.NewRichText(
		&widget.TextSegment{
			Text:  text,
			Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Monospace: true}},
		},
	)
	mono.Wrapping = fyne.TextWrapWord
	b.Mu.Lock()
	b.Content.Objects = []fyne.CanvasObject{mono}
	fyne.DoAndWait(func() {
		b.Title.SetText(scheme)
	})
	b.Mu.Unlock()
}

func pop(stack *[]string, name string) {
	st := *stack
	for i := len(st) - 1; i >= 0; i-- {
		if st[i] == name {
			*stack = st[:i]
			return
		}
	}
}

type toolbarLayout struct{}

func (t *toolbarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	btnSize := objects[2].MinSize()
	padding := theme.Padding()
	availableWidth := size.Width - btnSize.Width - (2 * padding)
	titleWidth := availableWidth / 4
	addressBarWidth := availableWidth * 3 / 4
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(titleWidth, size.Height))
	objects[1].Move(fyne.NewPos(titleWidth+padding, 0))
	objects[1].Resize(fyne.NewSize(addressBarWidth, size.Height))
	objects[2].Move(fyne.NewPos(titleWidth+padding+addressBarWidth+padding, 0))
	objects[2].Resize(fyne.NewSize(btnSize.Width, size.Height))
}

func (t *toolbarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	titleMin := objects[0].MinSize()
	entryMin := objects[1].MinSize()
	btnMin := objects[2].MinSize()
	minWidth := titleMin.Width + entryMin.Width + btnMin.Width + (2 * theme.Padding())
	minHeight := titleMin.Height
	if entryMin.Height > minHeight {
		minHeight = entryMin.Height
	}
	if btnMin.Height > minHeight {
		minHeight = btnMin.Height
	}

	return fyne.NewSize(minWidth, minHeight)
}

func newToolbarLayout() fyne.Layout {
	return &toolbarLayout{}
}

func (b *Browser) addToHistory(url string) {
	if len(b.History) == 0 || b.History[len(b.History)-1] != url {
		b.History = append(b.History, url)
		// Keep only latest 10
		if len(b.History) > 10 {
			b.History = b.History[len(b.History)-10:]
		}
		b.saveHistory()
	}
}

func (b *Browser) showHistory() {
	if len(b.History) == 0 {
		dialog.ShowInformation("History", "No history.", b.Window)
		return
	}
	history := make([]string, len(b.History))
	copy(history, b.History)
	for i := 0; i < len(history)/2; i++ {
		history[i], history[len(history)-1-i] = history[len(history)-1-i], history[i]
	}
	list := widget.NewList(
		func() int { return len(history) },
		func() fyne.CanvasObject { return widget.NewButton("", nil) },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			btn := o.(*widget.Button)
			entry := history[i]
			btn.SetText(entry)
			btn.OnTapped = func() {
				b.AddressBar.SetText(entry)
				go b.LoadAndRender(entry)
			}
		},
	)

	d := dialog.NewCustom("History", "Close",
		container.NewVScroll(list),
		b.Window,
	)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

func (b *Browser) saveHistory() {
	if len(b.History) == 0 {
		return
	}
	data, err := json.MarshalIndent(b.History, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling history:", err)
		return
	}
	err = os.WriteFile(b.HistoryFile, data, 0644)
	if err != nil {
		fmt.Println("Error saving history:", err)
	}
}

func (b *Browser) loadHistory() {
	data, err := os.ReadFile(b.HistoryFile)
	if err == nil && len(data) > 0 {
		json.Unmarshal(data, &b.History)
	}
}

func Run() {
	a := app.NewWithID("miki.browser")
	logo := assets.Get("logo.png")
	if logo != nil {
		a.SetIcon(logo)
	}
	b := NewBrowser(a)
	if len(os.Args) >= 2 {
		arg := os.Args[1]
		b.AddressBar.SetText(arg)
		go b.LoadAndRender(arg)
	}
	b.Window.ShowAndRun()
}
