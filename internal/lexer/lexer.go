package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TokenKind int

const (
	TextTok TokenKind = iota
	StartTagTok
	EndTagTok
)

type Token struct {
	Kind TokenKind
	Data string
}

func LexTokens(body string) []Token {
	var out []Token
	var text_buffer strings.Builder
	var tag_buffer strings.Builder
	inTag := false

	flushText := func() {
		if text_buffer.Len() == 0 {
			return
		}
		out = append(out, Token{Kind: TextTok, Data: HtmlUnescape(text_buffer.String())})
		text_buffer.Reset()
	}
	flushTag := func() {
		if tag_buffer.Len() == 0 {
			return
		}
		raw := strings.TrimSpace(tag_buffer.String())
		tag_buffer.Reset()
		if len(raw) == 0 || raw[0] == '!' || raw[0] == '?' {
			return
		}
		end := false
		if strings.HasPrefix(raw, "/") {
			end = true
			raw = strings.TrimSpace(raw[1:])
		}
		name := raw
		if sp := strings.IndexFunc(raw, unicode.IsSpace); sp >= 0 {
			name = raw[:sp]
		}
		name = strings.ToLower(name)
		if end {
			out = append(out, Token{Kind: EndTagTok, Data: name})
		} else {
			out = append(out, Token{Kind: StartTagTok, Data: name})
		}
	}

	for _, r := range body {
		switch r {
		case '<':
			if inTag {
				tag_buffer.WriteRune(r)
			} else {
				inTag = true
				flushText()
			}
		case '>':
			if inTag {
				inTag = false
				flushTag()
			} else {
				text_buffer.WriteRune(r)
			}
		default:
			if inTag {
				tag_buffer.WriteRune(r)
			} else {
				text_buffer.WriteRune(r)
			}
		}
	}
	if inTag {
		text_buffer.WriteString("<")
		text_buffer.WriteString(text_buffer.String())
	}
	flushText()
	fmt.Println(out)
	return out
}

func HtmlUnescape(s string) string {
	// minimal entity support
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#34;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&middot;", "·")
	s = strings.ReplaceAll(s, "&bull;", "•")
	s = strings.ReplaceAll(s, "&copy;", "©")
	s = strings.ReplaceAll(s, "&ndash;", "–")
	s = strings.ReplaceAll(s, "&mdash;", "—")
	return s
}

func ValidUTF_8(body []byte) (string, error) {
	var b strings.Builder
	for len(body) > 0 {
		r, size := utf8.DecodeRune(body)
		if r == utf8.RuneError && size == 1 {
			b.WriteRune('\uFFFD')
			body = body[1:]
		} else {
			b.WriteRune(r)
			body = body[size:]
		}
	}
	return b.String(), nil
}
