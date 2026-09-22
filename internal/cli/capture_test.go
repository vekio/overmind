package cli

import (
	"strings"
	"testing"
)

func TestCaptureTextUsesArgument(t *testing.T) {
	text, err := captureText(strings.NewReader("ignored"), []string{"quick note"})
	if err != nil {
		t.Fatalf("captureText() error = %v", err)
	}
	if text != "quick note" {
		t.Fatalf("captureText() = %q, want %q", text, "quick note")
	}
}

func TestCaptureTextReadsStandardInput(t *testing.T) {
	text, err := captureText(strings.NewReader("first line\nsecond line\n"), nil)
	if err != nil {
		t.Fatalf("captureText() error = %v", err)
	}
	if text != "first line\nsecond line" {
		t.Fatalf("captureText() = %q, want multiline input", text)
	}
}

func TestCaptureTextRejectsEmptyInput(t *testing.T) {
	if _, err := captureText(strings.NewReader("\n"), nil); err == nil {
		t.Fatal("captureText() error = nil, want empty input error")
	}
}
