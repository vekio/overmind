package cli

import (
	"testing"
)

func TestRebuildCommandPrintsIndexedCount(t *testing.T) {
	client := &otherCommandsClient{}
	got := runCLICommand(t, newRebuildCommand(fixedClient(client)), []string{"rebuild"}, "")
	if got != "indexed 3 note(s)\n" || client.rebuilds != 1 {
		t.Fatalf("rebuild output=%q calls=%d", got, client.rebuilds)
	}
}
