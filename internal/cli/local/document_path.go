package local

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"git.casta.me/alberto/overmind/internal/app/listdocuments"
	urfavecli "github.com/urfave/cli/v3"
)

// documentPath reads one logical path from the named argument or from the
// first TSV field on standard input.
func documentPath(command *urfavecli.Command, argumentName string) (string, error) {
	if paths := command.StringArgs(argumentName); len(paths) != 0 {
		return paths[0], nil
	}
	if isTerminal(command.Reader) {
		return "", fmt.Errorf("document path is required")
	}

	scanner := bufio.NewScanner(command.Reader)
	documentPath := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if documentPath != "" {
			return "", fmt.Errorf("standard input must contain exactly one document")
		}
		documentPath, _, _ = strings.Cut(line, "\t")
		documentPath = strings.TrimSpace(documentPath)
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read document path from standard input: %w", err)
	}
	if documentPath == "" {
		return "", fmt.Errorf("document path is required")
	}
	return documentPath, nil
}

func completeDocumentPaths(ctx context.Context, command *urfavecli.Command, state *applicationState) {
	application, err := state.Application()
	if err != nil {
		return
	}
	result, err := application.Queries.ListDocuments.Handle(ctx, listdocuments.ListDocumentsQuery{
		PathPrefix: command.Args().First(),
	})
	if err != nil {
		return
	}
	for _, document := range result.Documents {
		_, _ = fmt.Fprintln(command.Root().Writer, document.Path)
	}
}

func isTerminal(reader io.Reader) bool {
	file, ok := reader.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
