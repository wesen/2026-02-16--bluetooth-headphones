package presets

import (
	"context"
	"fmt"

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
	"soundctl/pkg/soundctl/preset"
)

// ── list ────────────────────────────────────────────────────────────────────

type listCommand struct {
	*cmds.CommandDescription
	store *preset.Store
}

func newListCommand(store *preset.Store) (*listCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &listCommand{
		CommandDescription: cmds.NewCommandDescription("list",
			cmds.WithShort("List saved presets"),
			cmds.WithSections(sections...),
		),
		store: store,
	}, nil
}

func (c *listCommand) RunIntoGlazeProcessor(ctx context.Context, _ *values.Values, gp middlewares.Processor) error {
	presets, err := c.store.List()
	if err != nil {
		return err
	}
	for _, p := range presets {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("name", p.Name),
			types.MRP("default_sink", p.DefaultSink),
			types.MRP("profiles", len(p.CardProfiles)),
			types.MRP("volumes", len(p.Volumes)),
			types.MRP("routes", len(p.AppRoutes)),
			types.MRP("updated", p.UpdatedAt.Format("2006-01-02 15:04")),
		)); err != nil {
			return err
		}
	}
	return nil
}

// ── apply ───────────────────────────────────────────────────────────────────

type applySettings struct {
	Name string `glazed:"name"`
}

type applyCommand struct {
	*cmds.CommandDescription
	store *preset.Store
	au    audio.Service
}

func newApplyCommand(store *preset.Store, au audio.Service) (*applyCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &applyCommand{
		CommandDescription: cmds.NewCommandDescription("apply",
			cmds.WithShort("Apply a saved preset"),
			cmds.WithFlags(
				fields.New("name", fields.TypeString, fields.WithRequired(true),
					fields.WithHelp("Preset name to apply")),
			),
			cmds.WithSections(sections...),
		),
		store: store,
		au:    au,
	}, nil
}

func (c *applyCommand) Run(ctx context.Context, vals *values.Values) error {
	s := &applySettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	p, err := c.store.Get(s.Name)
	if err != nil {
		return err
	}
	result := preset.Apply(ctx, c.au, p)
	for _, change := range result.Applied {
		fmt.Printf("  ✓ %s\n", change)
	}
	for _, e := range result.Errors {
		fmt.Printf("  ✗ %v\n", e)
	}
	if len(result.Errors) == 0 {
		fmt.Printf("Preset %q applied successfully.\n", p.Name)
	} else {
		fmt.Printf("Preset %q applied with %d error(s).\n", p.Name, len(result.Errors))
	}
	return nil
}

func (c *applyCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &applySettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	p, err := c.store.Get(s.Name)
	if err != nil {
		return err
	}
	result := preset.Apply(ctx, c.au, p)
	for _, change := range result.Applied {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("kind", "applied"),
			types.MRP("change", change),
		)); err != nil {
			return err
		}
	}
	for _, e := range result.Errors {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("kind", "error"),
			types.MRP("change", e.Error()),
		)); err != nil {
			return err
		}
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("kind", "summary"),
		types.MRP("preset", p.Name),
		types.MRP("applied", len(result.Applied)),
		types.MRP("errors", len(result.Errors)),
		types.MRP("ok", len(result.Errors) == 0),
	))
}

// ── save ────────────────────────────────────────────────────────────────────

type saveSettings struct {
	Name        string `glazed:"name"`
	DefaultSink string `glazed:"default-sink"`
}

type saveCommand struct {
	*cmds.CommandDescription
	store *preset.Store
}

func newSaveCommand(store *preset.Store) (*saveCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &saveCommand{
		CommandDescription: cmds.NewCommandDescription("save",
			cmds.WithShort("Save a preset (creates or updates)"),
			cmds.WithFlags(
				fields.New("name", fields.TypeString, fields.WithRequired(true),
					fields.WithHelp("Preset name")),
				fields.New("default-sink", fields.TypeString, fields.WithDefault(""),
					fields.WithHelp("Default sink name")),
			),
			cmds.WithSections(sections...),
		),
		store: store,
	}, nil
}

func (c *saveCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &saveSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	p := preset.Preset{
		Name:        s.Name,
		DefaultSink: s.DefaultSink,
	}
	if err := c.store.Save(p); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("operation", "presets.save"),
		types.MRP("name", p.Name),
		types.MRP("ok", true),
	))
}

// ── delete ──────────────────────────────────────────────────────────────────

type deleteSettings struct {
	Name string `glazed:"name"`
}

type deleteCommand struct {
	*cmds.CommandDescription
	store *preset.Store
}

func newDeleteCommand(store *preset.Store) (*deleteCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &deleteCommand{
		CommandDescription: cmds.NewCommandDescription("delete",
			cmds.WithShort("Delete a saved preset"),
			cmds.WithFlags(
				fields.New("name", fields.TypeString, fields.WithRequired(true),
					fields.WithHelp("Preset name to delete")),
			),
			cmds.WithSections(sections...),
		),
		store: store,
	}, nil
}

func (c *deleteCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &deleteSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	if err := c.store.Delete(s.Name); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("operation", "presets.delete"),
		types.MRP("name", s.Name),
		types.MRP("ok", true),
	))
}

// ── snapshot ────────────────────────────────────────────────────────────────

type snapshotSettings struct {
	Name string `glazed:"name"`
}

type snapshotCommand struct {
	*cmds.CommandDescription
	store *preset.Store
	au    audio.Service
}

func newSnapshotCommand(store *preset.Store, au audio.Service) (*snapshotCommand, error) {
	sections, err := common.DefaultSections()
	if err != nil {
		return nil, err
	}
	return &snapshotCommand{
		CommandDescription: cmds.NewCommandDescription("snapshot",
			cmds.WithShort("Capture current live state as a named preset"),
			cmds.WithFlags(
				fields.New("name", fields.TypeString, fields.WithRequired(true),
					fields.WithHelp("Name for the new preset")),
			),
			cmds.WithSections(sections...),
		),
		store: store,
		au:    au,
	}, nil
}

func (c *snapshotCommand) Run(ctx context.Context, vals *values.Values) error {
	s := &snapshotSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	p, err := preset.SnapshotCurrent(ctx, c.au)
	if err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}
	p.Name = s.Name
	if err := c.store.Save(p); err != nil {
		return err
	}
	fmt.Printf("Preset %q saved from current state.\n", p.Name)
	fmt.Printf("  Default sink: %s\n", p.DefaultSink)
	fmt.Printf("  Card profiles: %d\n", len(p.CardProfiles))
	fmt.Printf("  App routes: %d\n", len(p.AppRoutes))
	return nil
}

func (c *snapshotCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &snapshotSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "decode settings")
	}
	p, err := preset.SnapshotCurrent(ctx, c.au)
	if err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}
	p.Name = s.Name
	if err := c.store.Save(p); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("operation", "presets.snapshot"),
		types.MRP("name", p.Name),
		types.MRP("default_sink", p.DefaultSink),
		types.MRP("profiles", len(p.CardProfiles)),
		types.MRP("routes", len(p.AppRoutes)),
		types.MRP("ok", true),
	))
}

// ── Registration ────────────────────────────────────────────────────────────

func Register(parent *cobra.Command, store *preset.Store, au audio.Service) error {
	listCmd, err := newListCommand(store)
	if err != nil {
		return err
	}
	applyCmd, err := newApplyCommand(store, au)
	if err != nil {
		return err
	}
	saveCmd, err := newSaveCommand(store)
	if err != nil {
		return err
	}
	deleteCmd, err := newDeleteCommand(store)
	if err != nil {
		return err
	}
	snapshotCmd, err := newSnapshotCommand(store, au)
	if err != nil {
		return err
	}

	glazed := []cmds.Command{listCmd, saveCmd, deleteCmd}
	for _, command := range glazed {
		cobraCmd, err := common.BuildCobra(command)
		if err != nil {
			return err
		}
		parent.AddCommand(cobraCmd)
	}

	// Dual mode for apply and snapshot (normal + glaze)
	dual := []cmds.Command{applyCmd, snapshotCmd}
	for _, command := range dual {
		cobraCmd, err := common.BuildCobraDual(command)
		if err != nil {
			return err
		}
		parent.AddCommand(cobraCmd)
	}

	return nil
}
