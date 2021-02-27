package ui

import (
	"bytes"
	"strconv"
	"strings"
)

const (
	escape        = "\033["
	finalizer     = "m" // change color ansi code
	normal        = "0;"
	bold          = "1;"
	dim           = "2;"
	underline     = "4;"
	blink         = "5;"
	inverse       = "7;"
	strikethrough = "9;"

	// Reset is an ANSI sequence to reset colors to defaults
	Reset = "\033[39;49m"
	// ResetFg is an ANSI sequence to reset foreground color to default
	ResetFg = "\033[39m"
	// ResetBg is an ANSI sequence to reset background color to default
	ResetBg = "\033[49m"
)

var colorCode = map[bool]string{
	true:  "3",
	false: "4",
}

var brightColorCode = map[bool]string{
	true:  "9",
	false: "10",
}

var colorsTerm = map[string]string{
	"black":   "0",
	"red":     "1",
	"green":   "2",
	"yellow":  "3",
	"blue":    "4",
	"magenta": "5",
	"cyan":    "6",
	"white":   "7",
	"reset":   "9",
}

// ColorizerFunc colorizes a string
type ColorizerFunc func(string) string

// ColorFunc returns a function that sets the color for the passed string
func ColorFunc(color string) ColorizerFunc {
	if color == "" {
		return func(s string) string {
			return s
		}
	}
	code := ColorCode(color)
	return func(s string) string {
		return code + s
	}
}

func hasFg(style string) bool {
	return len(style) > 0 && strings.Index(style, ":") != 0
}

func hasBg(style string) bool {
	return len(style) > 0 && strings.Index(style, ":") == 0
}

// ColorWrapFunc returns a function that sets the color for the passed string and then resets it to default
func ColorWrapFunc(color string) ColorizerFunc {
	if color == "" {
		return func(s string) string {
			return s
		}
	}
	code := ColorCode(color)
	var resetCode string
	fg := hasFg(code)
	bg := hasBg(code)
	if fg && bg {
		resetCode = Reset
	} else if fg {
		resetCode = ResetFg
	} else {
		resetCode = ResetBg
	}
	return func(s string) string {
		return code + s + resetCode
	}
}

func parseAttrs(attrs string, front bool) string {
	if attrs == "" {
		return ""
	}
	attrGroup := strings.Split(attrs, "+")
	var color string
	if code, ok := colorsTerm[attrGroup[0]]; !ok {
		color = "9"
	} else {
		color = code
	}
	var style string
	if len(attrGroup) > 1 {
		style = attrGroup[1]
	}
	var colorBase string
	if strings.Contains(style, "h") {
		colorBase = brightColorCode[front]
	} else {
		colorBase = colorCode[front]
	}
	var buf = bytes.NewBufferString("")
	buf.WriteString(colorBase)
	buf.WriteString(color)
	buf.WriteString(";")
	if len(style) > 0 && front {
		if strings.Contains(style, "b") {
			buf.WriteString(bold)
		}
		if strings.Contains(style, "d") {
			buf.WriteString(dim)
		}
		if strings.Contains(style, "B") {
			buf.WriteString(blink)
		}
		if strings.Contains(style, "u") {
			buf.WriteString(underline)
		}
		if strings.Contains(style, "i") {
			buf.WriteString(inverse)
		}
		if strings.Contains(style, "s") {
			buf.WriteString(strikethrough)
		}
	}
	return buf.String()
}

// ColorCode returns the ANSI color color code for style.
func ColorCode(styleCode string) string {
	style := strings.ToLower(styleCode)
	if style == "reset" {
		return Reset
	}
	fgBg := strings.Split(style, ":")
	fgAttrs := parseAttrs(fgBg[0], true)
	var bgAttrs string
	if len(fgBg) > 1 {
		bgAttrs = parseAttrs(fgBg[1], false)
	}
	buf := bytes.NewBufferString(escape)
	buf.WriteString(fgAttrs)
	if bgAttrs != "" {
		buf.WriteString(bgAttrs)
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(finalizer)
	return buf.String()
}

// RGB creates int color from three components
func RGB(r, g, b uint) uint {
	return (r&0xFF)<<16 | (g&0xFF)<<8 | (b & 0xFF)
}

func rgbColor(color uint) string {
	var buf = bytes.NewBufferString("2;")
	r := (color >> 16) & 0xFF
	g := (color >> 8) & 0xFF
	b := color & 0xFF
	buf.WriteString(strconv.FormatInt(int64(r), 10))
	buf.WriteString(";")
	buf.WriteString(strconv.FormatInt(int64(g), 10))
	buf.WriteString(";")
	buf.WriteString(strconv.FormatInt(int64(b), 10))
	buf.WriteString(";")
	return buf.String()
}

// ColorRGB returns the ANSI color color for an RGB color
// pass -1 if you don't want to change either foreground or background
func ColorRGB(fg int, bg int) string {
	var buf = bytes.NewBufferString(escape)
	if fg >= 0 {
		buf.WriteString("38;")
		buf.WriteString(rgbColor(uint(fg)))
	}
	if bg >= 0 {
		buf.WriteString("48;")
		buf.WriteString(rgbColor(uint(bg)))
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteRune('m')
	return buf.String()
}
