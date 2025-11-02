package mtheme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type TemeVariant struct {
	fyne.Theme

	Variant fyne.ThemeVariant
}

func (f *TemeVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.Variant)
}

type MTheme struct {
	FontSize float32
}

func (t *MTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{20, 30, 50, 255}
	case theme.ColorNamePrimary:
		return color.RGBA{80, 130, 200, 255}
	case theme.ColorNameButton:
		return color.RGBA{80, 130, 200, 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *MTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *MTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *MTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return t.FontSize
	default:
		return theme.DefaultTheme().Size(name)
	}
}
