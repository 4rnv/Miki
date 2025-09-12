package main

// WILL BE CHANGED TO PACKAGE GUI AFTER TESTING

import (
	"fmt"
	"miki/internal/yurl"
	"os"
)

const WIDTH = 800
const HEIGHT = 600
const HSTEP, VSTEP = 12, 18
const SCROLL_STEP = 50

type Browser struct {
	Window       string // replace with fyne window
	Width        int
	Height       int
	Canvas       string // replace with fyne canvas
	Text         string
	Display_List []string
}

func Special_Page() {
	//draw special page here
}

func (Browser) Load(url yurl.URL, browser Browser) {
	content, err := url.Request(0)
	if err != nil {
		Special_Page()
		return
	}
	browser.Text = yurl.Lex(content)
	browser.Display_List = append(browser.Display_List, Layout(browser.Text, browser.Width))
	//Draw()
}

func Layout(text string, width int) string {
	return ""
}

func Draw() {
}

// WILL BE CHANGED TO FUNC RUN AFTER TESTING
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Pass URL as argument: go run main.go <url>")
		os.Exit(2)
	}
	raw := os.Args[1]
	u := yurl.NewURL(raw)
	content, err := u.Request(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(content)
}
