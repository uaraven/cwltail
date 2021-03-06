package ui

import (
	"bytes"
	"github.com/dlclark/regexp2"
	"regexp"
	"strings"
)

var (
	colors = []string{
		"green",
		"yellow",
		"blue",
		"magenta",
		"cyan",
		"red",
	}

	// errorBack = ColorFunc(":#FFD0D0")
	errorBack = ColorWrapFunc(":red")

	warnBack = ColorWrapFunc(":yellow")

	//StreamNameColorizer is a colorizer function for log stream name
	StreamNameColorizer = ColorWrapFunc("+b")
	//TimestampColorizer is a colorizer function for log event timestamp
	TimestampColorizer = ColorWrapFunc("+i")

	colorFuncs = createColorFuncs()

	standardColorsRe = regexp2.MustCompile(`(?<darkgoldenrod>\d{2}(?:[-:.]\d{2,3}){1,3})|`+
		`(?:\[(?<darkcyan>[^]]+)])|`+
		`(?:\s+(?<RoyalBlue>(?:"[^"]+")|(?:'[^']+'))\s+)|`+
		`(?:\b(?<plum>info|error|warn|trace|debug|warning)\b)|`+
		`(?:\b(?<bisque>\d[\d.]+)\b)|`+
		`(?:(?<turquoise>[\p{L}\d._]+)(?:=|:)(?<lightseagreen>['"]?[\p{L}\d._]+)['"]?)`, // key/value
		regexp2.IgnoreCase)
)

func createColorFuncs() []ColorizerFunc {
	result := make([]ColorizerFunc, len(colors))
	for i, color := range colors {
		result[i] = ColorWrapFunc(color)
	}
	return result
}

// HighlightLogLevel applies loglevel-specific background color to the provided text
// Warning background is applied if logLevel is equal to "WARN" or "WARNING" and
// Error background is applied for "ERROR" logLevel
func HighlightLogLevel(detectedLevels []string, matches []string, text string) string {
	for i, level := range detectedLevels {
		if matches[i] != "" {
			switch level {
			case "error":
				return errorBack(text)
			case "warning":
				return warnBack(text)
			}
		}
	}
	return text
}

// Colorize adds ansi color tags to each regex groups if the pattern matches
func Colorize(pattern *regexp.Regexp, text string) string {
	grpIndices := pattern.FindStringSubmatchIndex(text)
	if grpIndices == nil {
		return text
	}
	var sb strings.Builder
	colorIndex := 0
	groupPosIndex := 2
	pos := 0
	for groupPosIndex < len(grpIndices) {
		beforeGroup := text[pos:grpIndices[groupPosIndex]]
		pos = grpIndices[groupPosIndex+1]
		sb.WriteString(beforeGroup)
		sb.WriteString(colorFuncs[colorIndex](text[grpIndices[groupPosIndex]:grpIndices[groupPosIndex+1]]))
		groupPosIndex += 2
		colorIndex++
		if colorIndex >= len(colorFuncs) {
			colorIndex = 0
		}
	}
	lastGroup := text[pos:]
	sb.WriteString(lastGroup)

	return sb.String()
}

// HighlightSelection highlights slice of the string with indexes provided in selection with a style
func HighlightSelection(text string, selection []int, style string) string {
	buf := bytes.NewBufferString("")
	buf.WriteString(text[:selection[0]])
	wrap := ColorWrap(text[selection[0]:selection[1]], style)
	buf.WriteString(wrap)
	buf.WriteString(text[selection[1]:])
	s := buf.String()
	return s
}

func ColorizeByColorName(text string) string {
	matcher, err := standardColorsRe.FindStringMatch(text)
	if err != nil {
		panic(err)
	}
	pos := 0
	var sb strings.Builder

	for matcher != nil && len(matcher.Groups()) > 1 {
		for _, group := range matcher.Groups()[1:] {
			if len(group.Captures) > 0 {
				capt := group.Capture
				beforeGroup := text[pos:capt.Index]
				pos = capt.Index + capt.Length
				sb.WriteString(beforeGroup)

				colorName := group.Name
				colorId := -1
				if colorName != "" {
					colorId = NameToAnsi256(colorName)
				}
				if colorId >= 0 {
					sb.WriteString(Color256Wrap(capt.String(), colorName))
				} else {
					sb.WriteString(capt.String())
				}
			}
		}
		matcher, err = standardColorsRe.FindNextMatch(matcher)
		if err != nil {
			panic(err)
		}
	}
	lastGroup := text[pos:]
	sb.WriteString(lastGroup)

	return sb.String()
}
