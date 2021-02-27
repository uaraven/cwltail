package ui

import (
	"bytes"
	"fmt"
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
	return len(style) > 0 && strings.Index(style, ":") >= 0
}

func ansiCodesForWrapping(styleCode string) (string, string) {
	style := strings.Trim(styleCode, " ")
	if style == "" {
		return "", ""
	}
	code := ColorCode(style)
	var resetCode string
	fg := hasFg(style)
	bg := hasBg(style)
	if fg && bg {
		resetCode = Reset
	} else if fg {
		resetCode = ResetFg
	} else {
		resetCode = ResetBg
	}
	return code, resetCode
}

// ColorWrap sets the style for the passed string and then resets it to default
func ColorWrap(text string, style string) string {
	if style == "" {
		return text
	}
	code, resetCode := ansiCodesForWrapping(style)
	return code + text + resetCode
}

// ColorWrapFunc returns a function that sets the style for the passed string and then resets it to default
func ColorWrapFunc(style string) ColorizerFunc {
	if style == "" {
		return func(s string) string {
			return s
		}
	}
	code, resetCode := ansiCodesForWrapping(style)
	return func(s string) string {
		return code + s + resetCode
	}
}

func parseAttrs(attrs string, front bool) string {
	if attrs == "" {
		return ""
	}
	var buf = bytes.NewBufferString("")
	if attrs[0] == '#' {
		if front {
			buf.WriteString("38;")
		} else {
			buf.WriteString("48;")
		}
		buf.WriteString(colorRGB(int(hexToRGB(attrs))))
		return buf.String()
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
	if len(styleCode) == 0 {
		return Reset
	}
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

// hexToRGB converts color string in #RRGGBB format into an integer
// if string doesn't represent a valid hex number this function will panic
func hexToRGB(hexColor string) uint {
	if len(hexColor) != 7 {
		panic(fmt.Errorf("Invalid color string: %v", hexColor))
	}
	color, err := strconv.ParseUint(hexColor[1:], 16, 32)
	if err != nil {
		panic(err)
	}
	return uint(color)
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

func colorRGB(col int) string {
	var buf = bytes.NewBufferString("")
	buf.WriteString(rgbColor(uint(col)))
	return buf.String()
}

// ColorRGB returns the ANSI color code for an RGB color
// pass -1 if you don't want to change either foreground or background
func ColorRGB(fg int, bg int) string {
	var buf = bytes.NewBufferString(escape)
	if fg >= 0 {
		buf.WriteString("38;")
		buf.WriteString(colorRGB(fg))
	}
	if bg >= 0 {
		buf.WriteString("48;")
		buf.WriteString(colorRGB(bg))
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteRune('m')
	return buf.String()
}
