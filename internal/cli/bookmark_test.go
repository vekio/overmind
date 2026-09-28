package cli

import (
	"testing"
)

func TestBookmarkCommandPassesURLAndPrintsPath(t *testing.T) {
	client := &otherCommandsClient{}
	got := runCLICommand(t, newBookmarkCommand(fixedClient(client)), []string{"bookmark", "--tag", "Reading", "https://example.com"}, "")
	if got != "/vault/bookmark.adoc\n" || client.bookmarkURL.String() != "https://example.com" {
		t.Fatalf("bookmark output=%q URL=%q", got, client.bookmarkURL)
	}
}
