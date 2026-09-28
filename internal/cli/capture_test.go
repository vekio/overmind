package cli

import (
	"io"
	"strings"
	"testing"
)

func TestCaptureTextUsesArgumentOrPipedInput(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		read io.Reader
		want string
	}{
		{"argument", []string{"first\nsecond"}, strings.NewReader("ignored"), "first\nsecond"},
		{"stdin", nil, strings.NewReader("first\nsecond\n\n"), "first\nsecond"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := captureText(tc.read, tc.args)
			if err != nil || got != tc.want {
				t.Fatalf("captureText() = %q, %v", got, err)
			}
		})
	}
	for _, args := range [][]string{nil, {"  "}} {
		if _, err := captureText(strings.NewReader("\n  \n"), args); err == nil {
			t.Fatalf("empty capture accepted with args %v", args)
		}
	}
	if _, err := captureText(nil, nil); err == nil {
		t.Fatal("missing capture input accepted")
	}
}

func TestCaptureCommandReadsPipedTextAndPrintsPath(t *testing.T) {
	client := &otherCommandsClient{}
	got := runCLICommand(t, newCaptureCommand(fixedClient(client)), []string{"capture"}, "piped text\n")
	if got != "/vault/inbox.adoc\n" || client.captureText != "piped text" {
		t.Fatalf("capture output=%q text=%q", got, client.captureText)
	}
}
