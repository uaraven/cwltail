package ui

import (
	"testing"
)

func TestHighlightSelection(t *testing.T) {
	actual := HighlightSelection("Text to color", []int{5, 7}, "red")
	expected := "Text \033[31mto\033[39m color"

	if actual != expected {
		t.Errorf("Expected: Text \\033[31mto\\033[39m color\n  Actual: %v", actual)
	}
}

func TestStandardHighlightSelection(t *testing.T) {
	actual := ColorizeStandard("10:12 INFO test test test 1 ")
	expected := "\033[32m10:12\033[39m \033[33mINFO\033[39m test test test \033[34m1\033[39m "

	if actual != expected {
		t.Errorf("\nExpected: %v\n  Actual: %v", expected, actual)
	}
}
