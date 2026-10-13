package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type colorMode uint8

const (
	colorOff colorMode = iota
	colorANSI
	color256
	colorTrue
)

type presentation struct {
	mode  colorMode
	light bool
	utf8  bool
}

func presentationFor(w io.Writer) presentation {
	p := presentation{utf8: localeUsesUTF8()}
	f, ok := w.(*os.File)
	if !ok {
		return p
	}
	info, err := f.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return p
	}
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled || os.Getenv("TERM") == "dumb" {
		return p
	}
	term := os.Getenv("TERM")
	colorTerm := strings.ToLower(os.Getenv("COLORTERM"))
	switch {
	case colorTerm == "truecolor" || colorTerm == "24bit":
		p.mode = colorTrue
	case strings.Contains(term, "256color"):
		p.mode = color256
	default:
		p.mode = colorANSI
	}
	p.light = lightTerminalBackground(os.Getenv("COLORFGBG"))
	return p
}

func localeUsesUTF8() bool {
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		locale := strings.ToLower(os.Getenv(key))
		if locale != "" {
			return strings.Contains(locale, "utf-8") || strings.Contains(locale, "utf8")
		}
	}
	return false
}

func lightTerminalBackground(value string) bool {
	parts := strings.Split(value, ";")
	if len(parts) == 0 {
		return false
	}
	background, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return false
	}
	return background == 7 || background == 11 || background == 15
}

func (p presentation) code(color string) string {
	if p.mode == colorOff {
		return ""
	}
	if p.mode == colorTrue {
		if p.light {
			switch color {
			case "iris":
				return "\x1b[38;2;88;71;155m"
			case "mint":
				return "\x1b[38;2;48;105;72m"
			case "peach":
				return "\x1b[38;2;147;67;42m"
			}
		}
		switch color {
		case "iris":
			return "\x1b[38;2;195;181;245m"
		case "mint":
			return "\x1b[38;2;188;228;204m"
		case "peach":
			return "\x1b[38;2;242;185;165m"
		case "badge":
			return "\x1b[48;2;167;156;232m\x1b[38;2;38;59;102;1m"
		}
	}
	if p.mode == color256 {
		if p.light {
			switch color {
			case "iris":
				return "\x1b[38;5;54m"
			case "mint":
				return "\x1b[38;5;29m"
			case "peach":
				return "\x1b[38;5;130m"
			}
		}
		switch color {
		case "iris":
			return "\x1b[38;5;183m"
		case "mint":
			return "\x1b[38;5;151m"
		case "peach":
			return "\x1b[38;5;216m"
		case "badge":
			return "\x1b[48;5;183m\x1b[38;5;17;1m"
		}
	}
	if p.light {
		switch color {
		case "iris":
			return "\x1b[35m"
		case "mint":
			return "\x1b[32m"
		case "peach":
			return "\x1b[31m"
		}
	}
	switch color {
	case "iris":
		return "\x1b[95m"
	case "mint":
		return "\x1b[92m"
	case "peach":
		return "\x1b[91m"
	case "badge":
		return "\x1b[105m\x1b[30;1m"
	}
	return ""
}

func (p presentation) paint(value, color string) string {
	code := p.code(color)
	if code == "" {
		return value
	}
	return code + value + "\x1b[0m"
}

func (p presentation) badge() string {
	label := "koru"
	if p.mode == colorOff {
		return label
	}
	return p.code("badge") + " " + label + " \x1b[0m"
}

func (p presentation) brandLine() string {
	separator := " | "
	if p.utf8 {
		separator = " · "
	}
	return p.badge() + separator + p.paint("v"+version, "iris")
}

func writeStatus(w io.Writer, label, value, color string) error {
	p := presentationFor(w)
	_, err := fmt.Fprintf(w, "%s %s\n", p.paint(label, color), value)
	return err
}
