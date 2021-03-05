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

func TestColorizeByColorName(t *testing.T) {
	log := "[test1] [test2] 11:12"
	actual := ColorizeByColorName(log)
	expected := "[\033[38;5;36mtest1\033[39;49m] [\033[38;5;36mtest2\033[39;49m] \033[38;5;136m11:12\033[39;49m"

	if actual != expected {
		t.Errorf("\nExpected: '%v'\n  Actual: '%v'", expected, actual)
	}
}
