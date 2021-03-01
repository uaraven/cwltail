package ui

import (
	"bytes"
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

	reset = ColorCode("reset")

	//StreamNameColorizer is a colorizer function for log stream name
	StreamNameColorizer = ColorWrapFunc("+b")
	//TimestampColorizer is a colorizer function for log event timestamp
	TimestampColorizer = ColorWrapFunc("+i")

	colorFuncs = createColorFuncs()

	standartColorsRegex = regexp.MustCompile(`(?i)(?:\b(\d[\d:.T-]+)\b)|(?:\[(.*)\])|(?:\s+(\d+)\s+)|(?:\b(info|error|warn)\b)`)
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
// Error backgroud is applied for "ERROR" logLevel
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

// Colorize adds cview color tags to each regex groups if the pattern matches
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

func ColorizeStandard(text string) string {
	grpIndices := standartColorsRegex.FindAllStringSubmatchIndex(text, -1)
	if grpIndices == nil {
		return text
	}
	var sb strings.Builder
	colorIndex := 0
	pos := 0
	for _, group := range grpIndices {
		beforeGroup := text[pos:group[0]]
		pos = group[1]
		sb.WriteString(beforeGroup)
		sb.WriteString(colorFuncs[colorIndex](text[group[0]:group[1]]))
		colorIndex++
		if colorIndex >= len(colorFuncs) {
			colorIndex = 0
		}
	}
	lastGroup := text[pos:]
	sb.WriteString(lastGroup)

	return sb.String()
}
