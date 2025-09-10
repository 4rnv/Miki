package yurl

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type URL struct {
	Raw      string
	Scheme   string
	Host     string
	Port     int
	Path     string
	Is_blank bool
}

func (_url URL) _url_() string {
	return string(_url.Scheme + " : " + _url.Host + ":" + strconv.Itoa(_url.Port) + "/" + _url.Path)
}

func NewURL(raw string) URL {
	var _url URL
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "about:blank" {
		_url.Scheme = "about"
		_url.Host = ""
		_url.Path = ""
		_url.Port = 0
		_url.Is_blank = true
		return _url
	}
	orig := raw
	var parts []string
	if strings.HasPrefix(orig, "view-source") {
		parts = strings.SplitN(orig, ":", 2)
		_url.Scheme = parts[0]
		rest := parts[1]
		if idx := strings.Index(rest, "/"); idx >= 0 {
			_url.Host = rest[:idx]
			_url.Path = rest[idx:]
		} else {
			_url.Host = rest
			_url.Path = "/"
		}
		_url.Port = 80
		return _url
	} else if strings.HasPrefix(orig, "data:") {
		parts = strings.SplitN(orig, ":", 2)
		_url.Scheme = parts[0]
		_url.Path = parts[1]
		return _url
	} else if strings.HasPrefix(orig, "file://") || strings.HasPrefix(orig, "file:") {
		_url.Scheme = "file"
		after := orig
		if strings.HasPrefix(after, "file://") {
			after = after[len("file://"):]
		} else {
			after = after[len("file:"):]
		}
		if !strings.HasPrefix(after, "/") {
			after = "/" + after
		}
		_url.Path = after
		return _url
	}
	if strings.Contains(orig, "://") {
		parsed, err := url.Parse(orig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[!] Malformed URL '%s' (%v), falling back to about:blank\n", orig, err)
			_url.Scheme = "about"
			_url.Is_blank = true
			return _url
		}
		_url.Scheme = parsed.Scheme
		_url.Path = parsed.RequestURI()
		host := parsed.Host
		if strings.Contains(host, ":") {
			h, p, err := net.SplitHostPort(host)
			if err == nil {
				_url.Host = h
				if pi, err2 := strconv.Atoi(p); err2 == nil {
					_url.Port = pi
				}
			} else {
				_url.Host = host
			}
		} else {
			_url.Host = host
		}
		if _url.Port == 0 {
			if _url.Scheme == "http" {
				_url.Port = 80
			} else if _url.Scheme == "https" {
				_url.Port = 443
			}
		}
	}

	fmt.Println("Printing from NewURL: ", _url)
	return _url
}

func (_url URL) Request(redirect_count int) (string, error) {
	fmt.Println("Printing from Request: ", _url)
	if _url.Is_blank || _url.Scheme == "about" {
		return "", nil
	}
	if redirect_count > 5 {
		fmt.Println("Too many redirects.")
		return "", errors.New("too many redirects")
	}

	if _url.Scheme == "data" {
		parts := strings.SplitN(_url.Path, ",", 2)
		if len(parts) == 2 {
			content_type := parts[0]
			content := parts[1]
			if content_type == "text/html" {
				return content, nil
			} else {
				return "", errors.New("unsupported data: content type")
			}
		}
		return "", errors.New("broken data: URL")
	}

	if _url.Scheme == "file" {
		path := _url.Path
		if runtime.GOOS == "windows" && len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
		path = filepath.Clean(path)
		fmt.Println("Just before reading:", path)
		filebytes, err := os.ReadFile(path)
		if err != nil {
			error_s := fmt.Errorf("file reading error: %w", err)
			return "", error_s
		}
		return string(filebytes), nil
	}

	if _url.Scheme == "view-source" {
		_url.Port = 80
		parts := strings.SplitN(_url._url_(), "/", 2)
		_url.Host, _url.Path = parts[0], parts[1]
		_url.Path = "/" + _url.Path
	}

	address := _url.Host
	if !strings.Contains(_url.Host, ":") && _url.Port != 0 {
		address = _url.Host + ":" + strconv.Itoa(_url.Port)
	}
	var conn net.Conn
	var err error
	if _url.Scheme == "https" {
		dialer := &net.Dialer{Timeout: 12 * time.Second}
		config := &tls.Config{ServerName: _url.Host}
		conn, err = tls.DialWithDialer(dialer, "tcp", address, config)
	} else {
		conn, err = net.DialTimeout("tcp", address, 12*time.Second)
	}
	if err != nil {
		return "", fmt.Errorf("error connecting: %w", err)
	}
	defer conn.Close()

	if _url.Path == "" {
		_url.Path = "/"
	}
	req := "GET " + _url.Path + " HTTP/1.1\r\n"
	req += "Connection: keep-alive\r\n"
	req += "User-Agent: AveMaria-Miki\r\n"
	req += "Host: " + _url.Host + "\r\n\r\n"

	n, err := conn.Write([]byte(req))
	fmt.Printf("Response size: %d", n)
	if err != nil {
		return "", fmt.Errorf("error writing to connection: %w", err)
	}
	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}
	statusLine = strings.TrimRight(statusLine, "\r\n")
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		fmt.Printf("Broken status line:\n %q", statusLine)
	}
	versionn := parts[0]
	status := parts[1]
	explanation := parts[2]
	fmt.Println("\n", strings.Repeat("* ", 8))
	fmt.Printf("%s %s %s \n", versionn, status, explanation)
	fmt.Println(strings.Repeat(" *", 8))
	headers := make(map[string]string)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading headers: %w", err)
		}
		if line == "\r\n" || line == "\n" {
			break
		}
		line = strings.TrimRight(line, "\r\n")
		colon := strings.Index(line, ":")
		if colon < 0 {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(line[:colon]))
		value := strings.TrimSpace(line[colon+1:])
		headers[name] = value
	}
	statusInt, _ := strconv.Atoi(status)
	if 300 <= statusInt && statusInt < 400 {
		location := headers["location"]
		if location != "" {
			var new_url string
			if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
				new_url = location
			} else {
				new_url = fmt.Sprintf("%s://%s%s", _url.Scheme, _url.Host, location)
			}
			fmt.Println("Redirecting to: ", new_url)
			return NewURL(new_url).Request(redirect_count + 1)
		}
	}
	content_length, _ := strconv.Atoi(headers["content-length"])
	body := make([]byte, content_length)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			remaining, _ := io.ReadAll(reader)
			body = append(body[:0], remaining...)
		} else {
			return "", fmt.Errorf("error reading body: %w", err)
		}
	}
	if !utf8.Valid(body) {
		body_string, err := ValidUTF_8(body)
		if err != nil {
			return Lex(string(body)), nil
		}
		return Lex(body_string), nil
	}
	return Lex(string(body)), nil
}

func Lex(body string) string {
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
