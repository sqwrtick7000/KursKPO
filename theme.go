package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// salaryTheme — кастомная тема с синим акцентом
type salaryTheme struct{}

func newSalaryTheme() fyne.Theme {
	return &salaryTheme{}
}

func (t *salaryTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 30, G: 100, B: 200, A: 255}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 30, G: 100, B: 200, A: 180}
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 0, G: 120, B: 215, A: 255}
	}
	return theme.DefaultTheme().Color(n, v)
}

func (t *salaryTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(s)
}

func (t *salaryTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (t *salaryTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNameText:
		return 14
	case theme.SizeNamePadding:
		return 6
	}
	return theme.DefaultTheme().Size(n)
}
