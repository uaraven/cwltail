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
