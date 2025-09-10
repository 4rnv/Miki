package main

import (
	"fmt"
	"miki/internal/yurl"
	"os"
)

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
