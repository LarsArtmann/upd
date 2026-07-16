package main

import (
	"image/color"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"github.com/LarsArtmann/upd"
)

// colorSchemeFunc returns a fang ColorSchemeFunc that honors upd's no-color
// flag. When cfg.NoColor is true (from -C/--no-color or UPD_NO_COLOR) fang
// renders help and errors without color styling. NO_COLOR and non-TTY stdout
// are already handled by fang's colorprofile writer.
func colorSchemeFunc(cfg *upd.Config) fang.ColorSchemeFunc {
	return func(lightDark lipgloss.LightDarkFunc) fang.ColorScheme {
		if cfg.NoColor {
			return noColorScheme()
		}

		return fang.DefaultColorScheme(lightDark)
	}
}

// noColorScheme returns a ColorScheme where every color is lipgloss.NoColor,
// so fang emits no ANSI escape codes for help, error, or man-page output.
func noColorScheme() fang.ColorScheme {
	return fang.ColorScheme{
		Base:           lipgloss.NoColor{},
		Title:          lipgloss.NoColor{},
		Description:    lipgloss.NoColor{},
		Codeblock:      lipgloss.NoColor{},
		Program:        lipgloss.NoColor{},
		DimmedArgument: lipgloss.NoColor{},
		Comment:        lipgloss.NoColor{},
		Flag:           lipgloss.NoColor{},
		FlagDefault:    lipgloss.NoColor{},
		Command:        lipgloss.NoColor{},
		QuotedString:   lipgloss.NoColor{},
		Argument:       lipgloss.NoColor{},
		Help:           lipgloss.NoColor{},
		Dash:           lipgloss.NoColor{},
		ErrorHeader:    [2]color.Color{lipgloss.NoColor{}, lipgloss.NoColor{}},
		ErrorDetails:   lipgloss.NoColor{},
	}
}
