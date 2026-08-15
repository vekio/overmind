package cli

import (
	"fmt"
	"os"
	"path/filepath"

	urfavecli "github.com/urfave/cli/v3"
)

func openAsciiDoc(command *urfavecli.Command, operation string) (string, *os.File, error) {
	path := command.StringArgs("path")[0]
	if filepath.Ext(path) != ".adoc" {
		return path, nil, fmt.Errorf("%s path %q must have the .adoc extension", operation, path)
	}
	file, err := os.Open(path)
	if err != nil {
		return path, nil, fmt.Errorf("open AsciiDoc file %q: %w", path, err)
	}
	return path, file, nil
}
