package cli

import (
	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/overmind/internal/domain"
)

func areaFlag(usage string) urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:    "area",
		Aliases: []string{"a"},
		Usage:   usage,
		Config:  urfavecli.StringConfig{TrimSpace: true},
	}
}

func tagFlag(usage string) urfavecli.Flag {
	return &urfavecli.StringSliceFlag{
		Name:    "tag",
		Aliases: []string{"t"},
		Usage:   usage,
	}
}

func parseTags(values []string) (domain.Tags, error) {
	tags := make([]domain.Tag, len(values))
	for index, value := range values {
		tag, err := domain.NewTag(value)
		if err != nil {
			return domain.Tags{}, err
		}
		tags[index] = tag
	}

	return domain.NewTags(tags...)
}
