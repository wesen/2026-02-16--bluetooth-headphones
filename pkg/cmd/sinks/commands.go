package sinks

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
	return &listCommand{CommandDescription: cmds.NewCommandDescription("list", cmds.WithShort("List output sinks"), cmds.WithSections(sections...)), svc: svc}, nil
}

func (c *listCommand) RunIntoGlazeProcessor(ctx context.Context, _ *values.Values, gp middlewares.Processor) error {
	sinks, err := c.svc.ListSinks(ctx)
	if err != nil {
		return err
	}
	for _, sink := range sinks {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("id", sink.ID),
			types.MRP("name", sink.Name),
			types.MRP("driver", sink.Driver),
			types.MRP("sample_spec", sink.SampleSpec),
			types.MRP("state", sink.State),
		)); err != nil {
			return err
		}
	}
	return nil
}

type setDefaultSettings struct {
	Sink string `glazed:"sink"`
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
			cmds.WithShort("Set default sink"),
			cmds.WithFlags(fields.New("sink", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Sink name or ID"))),
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
	if err := c.svc.SetDefaultSink(ctx, s.Sink); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("operation", "sinks.set-default"), types.MRP("sink", s.Sink), types.MRP("ok", true)))
}

type moveSettings struct {
	StreamID int    `glazed:"stream-id"`
	Sink     string `glazed:"sink"`
}

type moveCommand struct {
	*cmds.CommandDescription
	svc audio.Service
}

func newMoveCommand(svc audio.Service) (*moveCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &moveCommand{
		CommandDescription: cmds.NewCommandDescription(
			"move-stream",
			cmds.WithShort("Move sink input stream to another sink"),
			cmds.WithFlags(
				fields.New("stream-id", fields.TypeInteger, fields.WithRequired(true), fields.WithHelp("Sink input stream ID")),
				fields.New("sink", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Sink name or ID")),
			),
			cmds.WithSections(sections...),
		),
		svc: svc,
	}, nil
}

func (c *moveCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &moveSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	if err := c.svc.MoveSinkInput(ctx, s.StreamID, s.Sink); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("operation", "sinks.move-stream"), types.MRP("stream_id", s.StreamID), types.MRP("sink", s.Sink), types.MRP("ok", true)))
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
	moveCmd, err := newMoveCommand(svc)
	if err != nil {
		return err
	}
	for _, command := range []cmds.Command{listCmd, setDefaultCmd, moveCmd} {
		cobraCmd, err := common.BuildCobra(command)
		if err != nil {
			return err
		}
		parent.AddCommand(cobraCmd)
	}
	return nil
}
