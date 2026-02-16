package volume

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

type setSettings struct {
	Target  string `glazed:"target"`
	Name    string `glazed:"name"`
	Percent int    `glazed:"percent"`
}

type setCommand struct {
	*cmds.CommandDescription
	svc audio.Service
}

func newSetCommand(svc audio.Service) (*setCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &setCommand{
		CommandDescription: cmds.NewCommandDescription(
			"set",
			cmds.WithShort("Set sink/source volume percentage"),
			cmds.WithFlags(
				fields.New("target", fields.TypeString, fields.WithDefault("sink"), fields.WithHelp("Target type: sink or source")),
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Sink/source name or ID")),
				fields.New("percent", fields.TypeInteger, fields.WithRequired(true), fields.WithHelp("Volume percent (0-150)")),
			),
			cmds.WithSections(sections...),
		),
		svc: svc,
	}, nil
}

func (c *setCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &setSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	if err := c.svc.SetVolume(ctx, s.Target, s.Name, s.Percent); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("operation", "volume.set"),
		types.MRP("target", s.Target),
		types.MRP("name", s.Name),
		types.MRP("percent", s.Percent),
		types.MRP("ok", true),
	))
}

func Register(parent *cobra.Command, svc audio.Service) error {
	setCmd, err := newSetCommand(svc)
	if err != nil {
		return err
	}
	cobraCmd, err := common.BuildCobra(setCmd)
	if err != nil {
		return err
	}
	parent.AddCommand(cobraCmd)
	return nil
}
