package lexer

import (
	"strings"
	"unicode/utf8"
)

type Text struct {
	Text string
}

type Tag struct {
	Tag string
}

func Lex(body string) string {
	// inTag := false
	// var out []any
	// var buf strings.Builder

	// flushText := func() {
	// 	if s := buf.String(); s != "" {
	// 		s = strings.ReplaceAll(strings.ReplaceAll(s, "&lt;", "<"), "&gt;", ">")
	// 		out = append(out, &Text{Text: s})
	// 		buf.Reset()
	// 	}
	// }

	// for _, r := range body {
	// 	switch r {
	// 	case '<':
	// 		inTag = true
	// 		flushText()
	// 	case '>':
	// 		inTag = false
	// 		tagStr := buf.String()
	// 		out = append(out, &Tag{Tag: tagStr})
	// 		buf.Reset()
	// 	default:
	// 		buf.WriteRune(r)
	// 	}
	// }
	// if !inTag {
	// 	flushText()
	// }

	// return out
	in_tag := false
	var b strings.Builder
	var text_to_print string
	for _, c := range body {
		if c == '<' {
			in_tag = true
		} else if c == '>' {
			in_tag = false
		} else if in_tag == false {
			b.WriteRune(c)
		}
	}
	text_to_print = b.String()
	text_to_print = strings.ReplaceAll(strings.ReplaceAll(text_to_print, "&lt;", "<"), "&gt;", ">")
	return text_to_print
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
