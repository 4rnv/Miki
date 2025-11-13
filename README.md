# Miki

A rudimentary web browser built in Go with the Fyne GUI framework. Based on exercises from https://browser.engineering.

<img src="public/c.png" width="600"/><br>
## Features

- Supported URL Schemes: `http`, `https`, `file`, `data`, `view-source`. `about:blank` is used as fallback. `http` is default.
    - `file` reads local files, `data` reads text content embedded directly into the url, `view-source` displays the raw HTML code sent by the server.
- Redirect handling: Upto 5 redirects for 3xx codes.
- Error handling: 4xx and 5xx codes are handled with user feedback.
- Lexer: Custom lexer for tokenizing server response. Tokenization is done on the tags (no support for classes or ids).
- Basic HTML styling: `<b>`, `<strong>`, `<i>`, `<pre>`, `<code>`, `<em>`, `<h1>`, `<a>` tags are rendered with styling.
- Image rendering: Very basic, sizing not implemented yet.
- Browsing history: Stored as `history.json` in the program/application directory.
- Theming: Light, dark, custom.

## Installation and Usage

### Prerequisites
- Go required ([install](https://go.dev/doc/install))
- Fyne also requires a C compiler for development. The compiler choice will vary depending upon the operating system. For more details refer to [Fyne docs](https://docs.fyne.io/started/quick/#prerequisites)

### Install
- `git clone https://github.com/4rnv/Miki.git`
- `cd Miki`
- `go run cmd/miki/main.go`

Alternatively, you can build a binary using the `go build` command.
- `go build -ldflags="-H=windowsgui" -o bin/miki.exe ./cmd/miki` (Windows)
- `go build -o bin/miki ./cmd/miki` (MacOS and Linux)

## Future work
- Hyperlinks should take you to that page
- Ctrl+O to open local files (.txt at least)
- Support for compressed content
- Caching and reusing connections
- Javascript support