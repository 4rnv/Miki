package assets

import (
	"embed"
	"fmt"

	"fyne.io/fyne/v2"
)

//go:embed *.png
var FS embed.FS

func Get(name string) fyne.Resource {
	data, err := FS.ReadFile(name)
	if err != nil {
		fmt.Println("Failed to read embedded asset:", name, err)
		return nil
	}
	return fyne.NewStaticResource(name, data)
}
