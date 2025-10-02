package main

// WILL BE CHANGED TO PACKAGE GUI AFTER TESTING

import (
	"miki/internal/lexer"
	"miki/internal/yurl"
	"net/url"
	"os"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Browser struct {
	Window     fyne.Window
	AddressBar *widget.Entry
	Scroll     *container.Scroll
	LoadBtn    *widget.Button
	Content    *fyne.Container
	Title      *widget.Label
	Text       string
	Mu         sync.Mutex
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

	b := &Browser{
		Window:     win,
		AddressBar: urlEntry,
		Title:      title,
	}
	b.LoadBtn = widget.NewButtonWithIcon("Load", theme.SearchIcon(), func() { go b.LoadAndRender(urlEntry.Text) })
	toolbar := container.NewGridWithColumns(3,
		container.NewStack(title),
		urlEntry,
		b.LoadBtn,
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
	textOrHTML := lexer.LexTokens(body)
	b.Mu.Lock()
	b.Text = ""
	fyne.DoAndWait(func() {
		b.Title.SetText("Loading...")
	})
	b.Mu.Unlock()
	b.renderTokens(textOrHTML)
}

func (b *Browser) Special_Page(err error) {
	var objs []fyne.CanvasObject
	img := canvas.NewImageFromFile("assets/sayak.png")
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

	type styleState struct{ bold, italic, link, inTitle bool }
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
			beginInline()
			seg := &widget.TextSegment{Text: tok.Data}
			seg.Style = widget.RichTextStyleInline
			if state.bold && state.italic {
				seg.Style.TextStyle = fyne.TextStyle{Bold: true, Italic: true}
			} else if state.bold {
				seg.Style.TextStyle = fyne.TextStyle{Bold: true}
			} else if state.italic {
				seg.Style.TextStyle = fyne.TextStyle{Italic: true}
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
			case "title":
				state.inTitle = true
				stack = append(stack, tok.Data)
			case "b", "strong":
				state.bold = true
				stack = append(stack, tok.Data)
			case "i", "em":
				state.italic = true
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
			default:
				stack = append(stack, tok.Data)
			}

		case lexer.EndTagTok:
			switch tok.Data {
			case "title":
				state.inTitle = false
				pop(&stack, tok.Data)
			case "b", "strong":
				state.bold = false
				pop(&stack, tok.Data)
			case "i", "em":
				state.italic = false
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

func pop(stack *[]string, name string) {
	st := *stack
	for i := len(st) - 1; i >= 0; i-- {
		if st[i] == name {
			*stack = st[:i]
			return
		}
	}
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
