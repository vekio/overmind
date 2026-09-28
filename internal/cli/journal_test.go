package cli

import (
	"reflect"
	"testing"
)

func TestJournalCommandPassesTagsAndPrintsPath(t *testing.T) {
	client := &otherCommandsClient{}
	got := runCLICommand(t, newJournalCommand(fixedClient(client)), []string{"journal", "--tag", "Daily"}, "")
	if got != "/vault/journal.adoc\n" || !reflect.DeepEqual(client.journalTags.Strings(), []string{"daily"}) {
		t.Fatalf("journal output=%q tags=%v", got, client.journalTags.Strings())
	}
}
