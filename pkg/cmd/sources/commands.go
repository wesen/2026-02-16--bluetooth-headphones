package sources

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

type listCommand struct {
	*cmds.CommandDescription
	svc audio.Service
}

func newListCommand(svc audio.Service) (*listCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &listCommand{CommandDescription: cmds.NewCommandDescription("list", cmds.WithShort("List input sources"), cmds.WithSections(sections...)), svc: svc}, nil
}

func (c *listCommand) RunIntoGlazeProcessor(ctx context.Context, _ *values.Values, gp middlewares.Processor) error {
	sources, err := c.svc.ListSources(ctx)
	if err != nil {
		return err
	}
	for _, source := range sources {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("id", source.ID),
			types.MRP("name", source.Name),
			types.MRP("driver", source.Driver),
			types.MRP("sample_spec", source.SampleSpec),
			types.MRP("state", source.State),
		)); err != nil {
			return err
		}
	}
	return nil
}

type setDefaultSettings struct {
	Source string `glazed:"source"`
}

type setDefaultCommand struct {
	*cmds.CommandDescription
	svc audio.Service
}

func newSetDefaultCommand(svc audio.Service) (*setDefaultCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &setDefaultCommand{
		CommandDescription: cmds.NewCommandDescription(
			"set-default",
			cmds.WithShort("Set default source"),
			cmds.WithFlags(fields.New("source", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Source name or ID"))),
			cmds.WithSections(sections...),
		),
		svc: svc,
	}, nil
}

func (c *setDefaultCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &setDefaultSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	if err := c.svc.SetDefaultSource(ctx, s.Source); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("operation", "sources.set-default"), types.MRP("source", s.Source), types.MRP("ok", true)))
}

func Register(parent *cobra.Command, svc audio.Service) error {
	listCmd, err := newListCommand(svc)
	if err != nil {
		return err
	}
	setDefaultCmd, err := newSetDefaultCommand(svc)
	if err != nil {
		return err
	}
	for _, command := range []cmds.Command{listCmd, setDefaultCmd} {
		cobraCmd, err := common.BuildCobra(command)
		if err != nil {
			return err
		}
		parent.AddCommand(cobraCmd)
	}
	return nil
}
