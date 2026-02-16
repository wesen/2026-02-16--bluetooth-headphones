package profiles

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
	return &listCommand{CommandDescription: cmds.NewCommandDescription("list", cmds.WithShort("List cards/profiles (short cards view)"), cmds.WithSections(sections...)), svc: svc}, nil
}

func (c *listCommand) RunIntoGlazeProcessor(ctx context.Context, _ *values.Values, gp middlewares.Processor) error {
	cards, err := c.svc.ListCards(ctx)
	if err != nil {
		return err
	}
	for _, card := range cards {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("id", card.ID),
			types.MRP("name", card.Name),
			types.MRP("driver", card.Driver),
			types.MRP("state", card.State),
		)); err != nil {
			return err
		}
	}
	return nil
}

type setSettings struct {
	Card    string `glazed:"card"`
	Profile string `glazed:"profile"`
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
			cmds.WithShort("Set card profile"),
			cmds.WithFlags(
				fields.New("card", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Card name")),
				fields.New("profile", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Profile name")),
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
	if err := c.svc.SetCardProfile(ctx, s.Card, s.Profile); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("operation", "profiles.set"), types.MRP("card", s.Card), types.MRP("profile", s.Profile), types.MRP("ok", true)))
}

func Register(parent *cobra.Command, svc audio.Service) error {
	listCmd, err := newListCommand(svc)
	if err != nil {
		return err
	}
	setCmd, err := newSetCommand(svc)
	if err != nil {
		return err
	}
	for _, command := range []cmds.Command{listCmd, setCmd} {
		cobraCmd, err := common.BuildCobra(command)
		if err != nil {
			return err
		}
		parent.AddCommand(cobraCmd)
	}
	return nil
}
