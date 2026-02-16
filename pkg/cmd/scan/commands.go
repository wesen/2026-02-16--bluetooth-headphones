package scan

import (
	"context"
	"fmt"
	"strings"

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

type actionCommand struct {
	*cmds.CommandDescription
	svc       bluetooth.Service
	operation string
	normalMsg string
	run       func(context.Context, bluetooth.Service) error
}

func newActionCommand(name, short, operation, normalMsg string, svc bluetooth.Service, run func(context.Context, bluetooth.Service) error) (*actionCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &actionCommand{
		CommandDescription: cmds.NewCommandDescription(name, cmds.WithShort(short), cmds.WithSections(sections...)),
		svc:                svc,
		operation:          operation,
		normalMsg:          normalMsg,
		run:                run,
	}, nil
}

func (c *actionCommand) Run(ctx context.Context, _ *values.Values) error {
	if err := c.run(ctx, c.svc); err != nil {
		return err
	}
	fmt.Println(c.normalMsg)
	return nil
}

func (c *actionCommand) RunIntoGlazeProcessor(ctx context.Context, _ *values.Values, gp middlewares.Processor) error {
	if err := c.run(ctx, c.svc); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("operation", c.operation), types.MRP("ok", true), types.MRP("note", c.normalMsg)))
}

type discoverSettings struct {
	Wait       int    `glazed:"wait"`
	NameFilter string `glazed:"name-filter"`
}

type discoverCommand struct {
	*cmds.CommandDescription
	svc bluetooth.Service
}

func newDiscoverCommand(svc bluetooth.Service) (*discoverCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &discoverCommand{
		CommandDescription: cmds.NewCommandDescription(
			"discover",
			cmds.WithShort("Run timed scan and return discovered devices"),
			cmds.WithFlags(
				fields.New("wait", fields.TypeInteger, fields.WithDefault(8), fields.WithHelp("Scan duration in seconds")),
				fields.New("name-filter", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Optional case-insensitive name filter")),
			),
			cmds.WithSections(sections...),
		),
		svc: svc,
	}, nil
}

func (c *discoverCommand) Run(ctx context.Context, vals *values.Values) error {
	s := &discoverSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	found, err := c.scan(ctx, s)
	if err != nil {
		return err
	}
	fmt.Printf("Scanning for %ds... found %d device(s).\n", s.Wait, len(found))
	for _, d := range found {
		fmt.Printf("- %s  %s\n", d.Address, d.Name)
	}
	return nil
}

func (c *discoverCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &discoverSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	found, err := c.scan(ctx, s)
	if err != nil {
		return err
	}
	for _, d := range found {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("kind", "discovered"),
			types.MRP("address", d.Address),
			types.MRP("name", d.Name),
		)); err != nil {
			return err
		}
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("kind", "summary"),
		types.MRP("operation", "scan.discover"),
		types.MRP("wait", s.Wait),
		types.MRP("count", len(found)),
		types.MRP("ok", true),
	))
}

func (c *discoverCommand) scan(ctx context.Context, s *discoverSettings) ([]bluetooth.DiscoveredDevice, error) {
	wait := s.Wait
	if wait <= 0 {
		wait = 8
	}
	found, err := c.svc.Discover(ctx, wait)
	if err != nil {
		return nil, err
	}
	return filterDiscovered(found, s.NameFilter), nil
}

type pairSettings struct {
	Addr       string `glazed:"addr"`
	Trust      bool   `glazed:"trust"`
	Connect    bool   `glazed:"connect"`
	Wait       int    `glazed:"wait"`
	NameFilter string `glazed:"name-filter"`
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
			cmds.WithLong("If --wait is set, run timed discovery before pairing. If --addr is omitted, pairing target is chosen from discovered devices (optionally filtered by --name-filter)."),
			cmds.WithFlags(
				fields.New("addr", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Bluetooth MAC address")),
				fields.New("trust", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Trust after pair")),
				fields.New("connect", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Connect after pair")),
				fields.New("wait", fields.TypeInteger, fields.WithDefault(0), fields.WithHelp("Optional pre-pair scan duration in seconds")),
				fields.New("name-filter", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Optional filter for auto-selecting discovered device")),
			),
			cmds.WithSections(sections...),
		),
		svc: svc,
	}, nil
}

func (c *pairCommand) Run(ctx context.Context, vals *values.Values) error {
	s := &pairSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	address, found, err := c.resolveAddress(ctx, s)
	if err != nil {
		return err
	}
	if s.Wait > 0 {
		fmt.Printf("Scan wait %ds complete, found %d device(s).\n", s.Wait, len(found))
		for _, d := range found {
			fmt.Printf("- %s  %s\n", d.Address, d.Name)
		}
	}
	if err := c.svc.Pair(ctx, address); err != nil {
		return err
	}
	if s.Trust {
		if err := c.svc.Trust(ctx, address); err != nil {
			return err
		}
	}
	if s.Connect {
		if err := c.svc.Connect(ctx, address); err != nil {
			return err
		}
	}
	fmt.Printf("Pair flow complete for %s (trust=%t connect=%t)\n", address, s.Trust, s.Connect)
	return nil
}

func (c *pairCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &pairSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	address, found, err := c.resolveAddress(ctx, s)
	if err != nil {
		return err
	}
	for _, d := range found {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("kind", "discovered"),
			types.MRP("address", d.Address),
			types.MRP("name", d.Name),
		)); err != nil {
			return err
		}
	}
	if err := c.svc.Pair(ctx, address); err != nil {
		return err
	}
	if s.Trust {
		if err := c.svc.Trust(ctx, address); err != nil {
			return err
		}
	}
	if s.Connect {
		if err := c.svc.Connect(ctx, address); err != nil {
			return err
		}
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("kind", "summary"),
		types.MRP("operation", "scan.pair"),
		types.MRP("address", address),
		types.MRP("trusted", s.Trust),
		types.MRP("connected", s.Connect),
		types.MRP("ok", true),
	))
}

func (c *pairCommand) resolveAddress(ctx context.Context, s *pairSettings) (string, []bluetooth.DiscoveredDevice, error) {
	if s.Addr != "" {
		if s.Wait <= 0 {
			return s.Addr, nil, nil
		}
		found, err := c.svc.Discover(ctx, s.Wait)
		if err != nil {
			return "", nil, err
		}
		return s.Addr, filterDiscovered(found, s.NameFilter), nil
	}

	if s.Wait <= 0 {
		return "", nil, fmt.Errorf("either --addr or --wait must be provided")
	}
	found, err := c.svc.Discover(ctx, s.Wait)
	if err != nil {
		return "", nil, err
	}
	filtered := filterDiscovered(found, s.NameFilter)
	if len(filtered) == 0 {
		return "", filtered, fmt.Errorf("no devices discovered matching filter %q", s.NameFilter)
	}
	if len(filtered) > 1 {
		matches := make([]string, 0, len(filtered))
		for _, d := range filtered {
			matches = append(matches, fmt.Sprintf("%s (%s)", d.Address, d.Name))
		}
		return "", filtered, fmt.Errorf("multiple devices matched, specify --addr: %s", strings.Join(matches, ", "))
	}
	return filtered[0].Address, filtered, nil
}

func filterDiscovered(found []bluetooth.DiscoveredDevice, filter string) []bluetooth.DiscoveredDevice {
	if filter == "" {
		return found
	}
	needle := strings.ToLower(filter)
	filtered := make([]bluetooth.DiscoveredDevice, 0, len(found))
	for _, d := range found {
		if strings.Contains(strings.ToLower(d.Name), needle) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func Register(parent *cobra.Command, svc bluetooth.Service) error {
	startCmd, err := newActionCommand("start", "Start bluetooth scanning (best effort)", "scan.start", "Triggered scan start (best effort; use scan discover --wait N to verify findings).", svc, func(ctx context.Context, s bluetooth.Service) error {
		return s.StartScan(ctx)
	})
	if err != nil {
		return err
	}
	stopCmd, err := newActionCommand("stop", "Stop bluetooth scanning", "scan.stop", "Triggered scan stop.", svc, func(ctx context.Context, s bluetooth.Service) error {
		return s.StopScan(ctx)
	})
	if err != nil {
		return err
	}
	discoverCmd, err := newDiscoverCommand(svc)
	if err != nil {
		return err
	}
	pairCmd, err := newPairCommand(svc)
	if err != nil {
		return err
	}

	dual := []cmds.Command{startCmd, stopCmd, discoverCmd, pairCmd}
	for _, command := range dual {
		cobraCmd, err := common.BuildCobraDual(command)
		if err != nil {
			return err
		}
		parent.AddCommand(cobraCmd)
	}
	return nil
}
