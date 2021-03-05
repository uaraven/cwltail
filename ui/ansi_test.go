package ui

import (
	"testing"
)

func TestForegroundColor(t *testing.T) {
	col := ColorCode("red+bu")
	if col != "\033[31;1;4m" {
		t.Errorf("Expected: \\033[31;1;4m\n  Actual: %v", col)
	}
}

func TestFgResetBg(t *testing.T) {
	col := ColorCode("red+bu:reset")
	if col != "\033[31;1;4;49m" {
		t.Errorf("Expected: \\033[31;1;4;49m\n  Actual: %v", col)
	}
}

func TestBgResetFg(t *testing.T) {
	col := ColorCode("reset:blue")
	if col != "\033[39;44m" {
		t.Errorf("Expected: \\033[39;44m\n  Actual: %v", col)
	}
}

func TestBackgroundColor(t *testing.T) {
	col := ColorCode(":red+bu")
	if col != "\033[41m" {
		t.Errorf("Expected: \\033[41m\n  Actual: %v", col)
	}
}

func TestFgAndBgColor(t *testing.T) {
	col := ColorCode("red+bu:blue+h")
	if col != "\033[31;1;4;104m" {
		t.Errorf("Expected: \\033[31;1;4;104m\n  Actual: %v", col)
	}
}

func TestRgbColor(t *testing.T) {
	col := ColorRGB(0xFF1001, 0x0110FF)
	if col != "\033[38;2;255;16;1;48;2;1;16;255m" {
		t.Errorf("\nExpected: \033[38;2;255;16;1;48;2;1;16;255m\n  Actual: %v", col)
	}
}

func TestRgbCode(t *testing.T) {
	col := ColorCode("#FF1001:#0110FF")
	if col != "\033[38;2;255;16;1;48;2;1;16;255m" {
		t.Errorf("\nExpected: \033[38;2;255;16;1;48;2;1;16;255m\n  Actual: %v", col)
	}
}

func TestWrapColorizer(t *testing.T) {
	colorizer := ColorWrapFunc("black:yellow")
	actual := colorizer("test")
	expected := "\033[30;43mtest\033[39;49m"
	if actual != expected {
		t.Errorf("\nExpected: %v\n  Actual: %v", actual, expected)
	}
}

func TestNameToAnsi256(t *testing.T) {
	color := NameToAnsi256("Goldenrod")
	if color != 179 {
		t.Errorf("Invalid color index returned, expected 179, got %d", color)
	}
}

func TestColorize256(t *testing.T) {
	colorizer := ColorWrapFunc("RosyBrown:DodgerBlue")
	actual := colorizer("Test")
	expected := "\033[38;5;138;48;5;33mTest\033[39;49m"
	if actual != expected {
		t.Errorf("\nExpected: %v\n  Actual: %v", expected, actual)
	}
}
