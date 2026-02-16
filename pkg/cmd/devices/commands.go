package devices

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

type listCommand struct {
	*cmds.CommandDescription
	svc bluetooth.Service
}

func newListCommand(svc bluetooth.Service) (*listCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &listCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List paired/known bluetooth devices"),
			cmds.WithSections(sections...),
		),
		svc: svc,
	}, nil
}

func (c *listCommand) RunIntoGlazeProcessor(ctx context.Context, _ *values.Values, gp middlewares.Processor) error {
	devices, err := c.svc.ListDevices(ctx)
	if err != nil {
		return err
	}
	for _, d := range devices {
		if err := gp.AddRow(ctx, types.NewRow(types.MRP("address", d.Address), types.MRP("name", d.Name))); err != nil {
			return err
		}
	}
	return nil
}

type addrSettings struct {
	Addr string `glazed:"addr"`
}

type mutateCommand struct {
	*cmds.CommandDescription
	svc       bluetooth.Service
	operation string
	run       func(ctx context.Context, svc bluetooth.Service, addr string) error
}

func newAddrCommand(name string, short string, operation string, svc bluetooth.Service, run func(ctx context.Context, svc bluetooth.Service, addr string) error) (*mutateCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &mutateCommand{
		CommandDescription: cmds.NewCommandDescription(
			name,
			cmds.WithShort(short),
			cmds.WithFlags(fields.New("addr", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Bluetooth MAC address"))),
			cmds.WithSections(sections...),
		),
		svc:       svc,
		operation: operation,
		run:       run,
	}, nil
}

func (c *mutateCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &addrSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	if err := c.run(ctx, c.svc, s.Addr); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("operation", c.operation), types.MRP("address", s.Addr), types.MRP("ok", true)))
}

func Register(parent *cobra.Command, svc bluetooth.Service) error {
	commands := []cmds.Command{}

	listCmd, err := newListCommand(svc)
	if err != nil {
		return err
	}
	connectCmd, err := newAddrCommand("connect", "Connect bluetooth device", "devices.connect", svc, func(ctx context.Context, s bluetooth.Service, addr string) error {
		return s.Connect(ctx, addr)
	})
	if err != nil {
		return err
	}
	disconnectCmd, err := newAddrCommand("disconnect", "Disconnect bluetooth device", "devices.disconnect", svc, func(ctx context.Context, s bluetooth.Service, addr string) error {
		return s.Disconnect(ctx, addr)
	})
	if err != nil {
		return err
	}
	trustCmd, err := newAddrCommand("trust", "Trust bluetooth device", "devices.trust", svc, func(ctx context.Context, s bluetooth.Service, addr string) error {
		return s.Trust(ctx, addr)
	})
	if err != nil {
		return err
	}
	forgetCmd, err := newAddrCommand("forget", "Remove bluetooth device", "devices.forget", svc, func(ctx context.Context, s bluetooth.Service, addr string) error {
		return s.Remove(ctx, addr)
	})
	if err != nil {
		return err
	}
	commands = append(commands, listCmd, connectCmd, disconnectCmd, trustCmd, forgetCmd)

	for _, command := range commands {
		cobraCmd, err := common.BuildCobra(command)
		if err != nil {
			return err
		}
		parent.AddCommand(cobraCmd)
	}
	return nil
}
