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

var colorNames256 = map[string]string{
	"black (system)":    "0",
	"maroon (system)":   "1",
	"green (system)":    "2",
	"olive (system)":    "3",
	"navy (system)":     "4",
	"purple (system)":   "5",
	"teal (system)":     "6",
	"silver (system)":   "7",
	"grey (system)":     "8",
	"red (system)":      "9",
	"lime (system)":     "10",
	"yellow (system)":   "11",
	"blue (system)":     "12",
	"fuchsia (system)":  "13",
	"aqua (system)":     "14",
	"white (system)":    "15",
	"grey0":             "16",
	"navyblue":          "17",
	"darkblue":          "18",
	"blue":              "21",
	"darkgreen":         "22",
	"dodgerblue":        "33",
	"darkcyan":          "36",
	"lightseagreen":     "37",
	"deepskyblue":       "39",
	"darkturquoise":     "30",
	"turquoise":         "45",
	"green":             "46",
	"springgreen":       "48",
	"mediumspringgreen": "49",
	"cyan":              "51",
	"blueviolet":        "57",
	"grey37":            "59",
	"slateblue":         "61",
	"royalblue":         "63",
	"darkseagreen":      "108",
	"steelblue":         "67",
	"cornflowerblue":    "69",
	"cadetblue":         "72",
	"seagreen":          "78",
	"mediumturquoise":   "80",
	"darkslategray":     "87",
	"darkred":           "88",
	"darkmagenta":       "90",
	"darkviolet":        "92",
	"mediumpurple":      "97",
	"grey53":            "102",
	"lightslategrey":    "103",
	"lightslateblue":    "105",
	"lightskyblue":      "109",
	"palegreen":         "114",
	"skyblue":           "117",
	"chartreuse":        "118",
	"lightgreen":        "120",
	"aquamarine":        "122",
	"mediumvioletred":   "126",
	"purple":            "129",
	"indianred":         "131",
	"mediumorchid":      "134",
	"darkgoldenrod":     "136",
	"rosybrown":         "138",
	"grey63":            "139",
	"darkkhaki":         "143",
	"grey69":            "145",
	"lightsteelblue":    "146",
	"greenyellow":       "154",
	"paleturquoise":     "159",
	"orchid":            "170",
	"violet":            "177",
	"goldenrod":         "179",
	"tan":               "180",
	"thistle":           "182",
	"plum":              "183",
	"khaki":             "185",
	"lightyellow":       "187",
	"grey84":            "188",
	"darkolivegreen":    "192",
	"honeydew":          "194",
	"lightcyan":         "195",
	"red":               "196",
	"deeppink":          "199",
	"magenta":           "201",
	"orangered":         "202",
	"hotpink":           "205",
	"darkorange":        "208",
	"salmon":            "209",
	"lightcoral":        "210",
	"palevioletred":     "211",
	"orange":            "214",
	"sandybrown":        "215",
	"lightsalmon":       "216",
	"lightpink":         "217",
	"pink":              "218",
	"gold":              "220",
	"navajowhite":       "223",
	"mistyrose":         "224",
	"yellow":            "226",
	"lightgoldenrod":    "227",
	"wheat":             "229",
	"cornsilk":          "230",
	"grey100":           "231",
	"grey3":             "232",
	"grey7":             "233",
	"grey11":            "234",
	"grey15":            "235",
	"grey19":            "236",
	"grey23":            "237",
	"grey27":            "238",
	"grey30":            "239",
	"grey35":            "240",
	"grey39":            "241",
	"grey42":            "242",
	"grey46":            "243",
	"grey50":            "244",
	"grey54":            "245",
	"grey58":            "246",
	"grey62":            "247",
	"grey66":            "248",
	"grey70":            "249",
	"grey74":            "250",
	"grey78":            "251",
	"grey82":            "252",
	"grey85":            "253",
	"grey89":            "254",
	"grey93":            "255",
}

// ColorizerFunc colorizes a string
type ColorizerFunc func(string) string

func NameToAnsi256(name string) int {
	name = strings.ToLower(name)
	var color string
	var ok bool
	if color, ok = colorNames256[name]; !ok {
		if color, ok = colorNames256[name+"(system)"]; !ok {
			color = "-1"
		}
	}
	c, _ := strconv.ParseInt(color, 10, 32)
	return int(c)
}

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

func Color256Wrap(text string, style string) string {
	colors = strings.Split(style, ":")
	fgColor := NameToAnsi256(colors[0])
	bgColor := -1
	if len(colors) > 1 {
		bgColor = NameToAnsi256(colors[1])
	}
	if fgColor < 0 && bgColor < 0 {
		return Reset + text
	}
	result := bytes.NewBufferString("\033[")
	if fgColor >= 0 {
		result.WriteString("38;5;" + strconv.Itoa(fgColor) + ";")
	}
	if bgColor >= 0 {
		result.WriteString("48;5;" + strconv.Itoa(bgColor) + ";")
	}
	result.Truncate(result.Len() - 1)
	result.WriteString("m")
	result.WriteString(text)
	result.WriteString(Reset)
	return result.String()
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
	var c256 bool
	attrGroup := strings.Split(attrs, "+")
	var color string
	code, ok := colorsTerm[attrGroup[0]]
	if !ok {
		if code, ok = colorNames256[attrGroup[0]]; !ok {
			color = "9"
		} else {
			color = "5;" + code
			c256 = true
		}
	} else {
		color = code
	}
	var style string
	if len(attrGroup) > 1 && !c256 {
		style = attrGroup[1]
	}
	var colorBase string
	if c256 {
		if front {
			colorBase = "38;"
		} else {
			colorBase = "48;"
		}
	} else if strings.Contains(style, "h") {
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

func Color256Code(styleCode string) string {
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
