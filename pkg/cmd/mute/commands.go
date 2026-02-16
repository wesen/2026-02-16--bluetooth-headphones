package mute

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"soundctl/pkg/cmd/common"
	"soundctl/pkg/soundctl/audio"
)

type toggleSettings struct {
	Target string `glazed:"target"`
	Name   string `glazed:"name"`
}

type toggleCommand struct {
	*cmds.CommandDescription
	svc audio.Service
}

func newToggleCommand(svc audio.Service) (*toggleCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &toggleCommand{
		CommandDescription: cmds.NewCommandDescription(
			"toggle",
			cmds.WithShort("Toggle sink/source mute"),
			cmds.WithFlags(
				fields.New("target", fields.TypeString, fields.WithDefault("sink"), fields.WithHelp("Target type: sink or source")),
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Sink/source name or ID")),
			),
			cmds.WithSections(sections...),
		),
		svc: svc,
	}, nil
}

func (c *toggleCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &toggleSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	if err := c.svc.ToggleMute(ctx, s.Target, s.Name); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("operation", "mute.toggle"),
		types.MRP("target", s.Target),
		types.MRP("name", s.Name),
		types.MRP("ok", true),
	))
}

func Register(parent *cobra.Command, svc audio.Service) error {
	toggleCmd, err := newToggleCommand(svc)
	if err != nil {
		return err
	}
	cobraCmd, err := common.BuildCobra(toggleCmd)
	if err != nil {
		return err
	}
	parent.AddCommand(cobraCmd)
	return nil
}
