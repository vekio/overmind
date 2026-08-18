package server

import (
	"context"
	"fmt"
	"net"
	"net/url"

	overmindhttp "git.casta.me/alberto/overmind/internal/http"
	urfavecli "github.com/urfave/cli/v3"
)

func newServeCommand(state *applicationState) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "serve",
		Usage: "serve the Overmind HTTP API",
		Flags: []urfavecli.Flag{addressFlag()},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			address := command.String("address")
			if address == "" {
				address = state.Config().HTTP.Address
			}

			logger, err := state.Logger()
			if err != nil {
				return err
			}
			server, err := overmindhttp.NewServer(address, logger)
			if err != nil {
				return err
			}
			return server.Serve(ctx, func(listenAddress string) error {
				localURL := localHTTPURL(listenAddress)
				_, err := fmt.Fprintf(command.Writer,
					"\n  Overmind is ready\n\n  ➜  Local:  %s\n  ➜  Health: %shealth\n\n",
					localURL,
					localURL,
				)
				return err
			})
		},
	}
}

func localHTTPURL(address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "http://" + address + "/"
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "localhost"
	}
	return (&url.URL{Scheme: "http", Host: net.JoinHostPort(host, port), Path: "/"}).String()
}
