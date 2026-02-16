package scan

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
	"soundctl/pkg/soundctl/bluetooth"
)

type simpleCommand struct {
	*cmds.CommandDescription
	svc       bluetooth.Service
	operation string
	run       func(context.Context, bluetooth.Service) error
}

func newSimpleCommand(name, short, operation string, svc bluetooth.Service, run func(context.Context, bluetooth.Service) error) (*simpleCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &simpleCommand{
		CommandDescription: cmds.NewCommandDescription(name, cmds.WithShort(short), cmds.WithSections(sections...)),
		svc:                svc,
		operation:          operation,
		run:                run,
	}, nil
}

func (c *simpleCommand) RunIntoGlazeProcessor(ctx context.Context, _ *values.Values, gp middlewares.Processor) error {
	if err := c.run(ctx, c.svc); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("operation", c.operation), types.MRP("ok", true)))
}

type pairSettings struct {
	Addr    string `glazed:"addr"`
	Trust   bool   `glazed:"trust"`
	Connect bool   `glazed:"connect"`
}

type pairCommand struct {
	*cmds.CommandDescription
	svc bluetooth.Service
}

func newPairCommand(svc bluetooth.Service) (*pairCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &pairCommand{
		CommandDescription: cmds.NewCommandDescription(
			"pair",
			cmds.WithShort("Pair a bluetooth device and optionally trust/connect"),
			cmds.WithFlags(
				fields.New("addr", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Bluetooth MAC address")),
				fields.New("trust", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Trust after pair")),
				fields.New("connect", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Connect after pair")),
			),
			cmds.WithSections(sections...),
		),
		svc: svc,
	}, nil
}

func (c *pairCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &pairSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	if err := c.svc.Pair(ctx, s.Addr); err != nil {
		return err
	}
	if s.Trust {
		if err := c.svc.Trust(ctx, s.Addr); err != nil {
			return err
		}
	}
	if s.Connect {
		if err := c.svc.Connect(ctx, s.Addr); err != nil {
			return err
		}
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("operation", "scan.pair"),
		types.MRP("address", s.Addr),
		types.MRP("trusted", s.Trust),
		types.MRP("connected", s.Connect),
		types.MRP("ok", true),
	))
}

func Register(parent *cobra.Command, svc bluetooth.Service) error {
	startCmd, err := newSimpleCommand("start", "Start bluetooth scanning", "scan.start", svc, func(ctx context.Context, s bluetooth.Service) error {
		return s.StartScan(ctx)
	})
	if err != nil {
		return err
	}
	stopCmd, err := newSimpleCommand("stop", "Stop bluetooth scanning", "scan.stop", svc, func(ctx context.Context, s bluetooth.Service) error {
		return s.StopScan(ctx)
	})
	if err != nil {
		return err
	}
	pairCmd, err := newPairCommand(svc)
	if err != nil {
		return err
	}

	for _, command := range []cmds.Command{startCmd, stopCmd, pairCmd} {
		cobraCmd, err := common.BuildCobra(command)
		if err != nil {
			return err
		}
		parent.AddCommand(cobraCmd)
	}
	return nil
}
