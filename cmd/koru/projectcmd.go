package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/plugins"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// builtinPluginOrder lists the integrated plugins in the order koru documents them.
var builtinPluginOrder = []string{"device", "storage", "vibration", "app", "share", "notifications"}

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

func newBumpCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "bump [patch|minor|major|code] [project]",
		Short:   "Increase versionCode and the version in koru.json",
		Long:    "Increase versionCode by one and, unless the level is \"code\", the semantic version X.Y.Z in koru.json.",
		Example: "  koru bump\n  koru bump minor ./my-app\n  koru bump code",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			level, project := "patch", "."
			switch {
			case len(args) == 2:
				level, project = args[0], args[1]
			case len(args) == 1 && isBumpLevel(args[0]):
				level = args[0]
			case len(args) == 1:
				project = args[0]
			}
			return bumpVersion(project, level, cmd.OutOrStdout())
		},
	}
}

func isBumpLevel(s string) bool {
	return s == "patch" || s == "minor" || s == "major" || s == "code"
}

func bumpVersion(project, level string, out io.Writer) error {
	if !isBumpLevel(level) {
		return fmt.Errorf("unknown bump level %q; use patch, minor, major or code", level)
	}
	c, err := config.Load(project)
	if err != nil {
		return err
	}
	oldVersion, oldCode := c.Version, c.VersionCode
	if c.VersionCode >= 2100000000 {
		return errors.New("versionCode is already at the Android limit 2100000000; it cannot be increased")
	}
	c.VersionCode++
	var warning string
	if level != "code" {
		if next, ok := bumpSemver(c.Version, level); ok {
			c.Version = next
		} else {
			warning = fmt.Sprintf("version %q is not X.Y.Z; only versionCode was increased. Edit version in koru.json by hand.", c.Version)
		}
	}
	if err = saveConfig(project, c); err != nil {
		return err
	}
	if warning != "" {
		if err = writeStatus(out, "Warning", warning, "peach"); err != nil {
			return err
		}
	}
	return writeStatus(out, "Bumped", fmt.Sprintf("version %s -> %s, versionCode %d -> %d", oldVersion, c.Version, oldCode, c.VersionCode), "mint")
}

func bumpSemver(v, level string) (string, bool) {
	m := semverPattern.FindStringSubmatch(v)
	if m == nil {
		return v, false
	}
	var n [3]uint64
	for i := range n {
		x, err := strconv.ParseUint(m[i+1], 10, 32)
		if err != nil {
			return v, false
		}
		n[i] = x
	}
	switch level {
	case "major":
		n = [3]uint64{n[0] + 1, 0, 0}
	case "minor":
		n = [3]uint64{n[0], n[1] + 1, 0}
	default:
		n[2]++
	}
	return fmt.Sprintf("%d.%d.%d", n[0], n[1], n[2]), true
}

// saveConfig validates c and rewrites koru.json atomically, keeping Config field order.
func saveConfig(project string, c config.Config) error {
	if c.Permissions == nil {
		c.Permissions = []string{}
	}
	if c.Plugins == nil {
		c.Plugins = []string{}
	}
	if err := c.Validate(); err != nil {
		return fmt.Errorf("refusing to write koru.json: %w", err)
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(project, "koru.json"), append(b, '\n'))
}

func newPluginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin",
		Short:   "List, enable and disable plugins in koru.json",
		Example: "  koru plugin list\n  koru plugin add vibration share\n  koru plugin remove share --project ./my-app",
	}
	listCmd := &cobra.Command{
		Use:     "list [project]",
		Short:   "Show integrated plugins, which are enabled and the permissions they add",
		Example: "  koru plugin list ./my-app",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := "."
			if len(args) == 1 {
				project = args[0]
			}
			return pluginList(project, cmd.OutOrStdout())
		},
	}
	addCmd := &cobra.Command{
		Use:     "add <id>...",
		Short:   "Enable plugins in koru.json",
		Example: "  koru plugin add vibration share --project ./my-app",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, _ := cmd.Flags().GetString("project")
			return pluginEdit(project, args, true, cmd.OutOrStdout())
		},
	}
	removeCmd := &cobra.Command{
		Use:     "remove <id>...",
		Short:   "Disable plugins in koru.json",
		Example: "  koru plugin remove share --project ./my-app",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, _ := cmd.Flags().GetString("project")
			return pluginEdit(project, args, false, cmd.OutOrStdout())
		},
	}
	for _, c := range []*cobra.Command{addCmd, removeCmd} {
		c.Flags().String("project", ".", "project directory containing koru.json")
	}
	cmd.AddCommand(listCmd, addCmd, removeCmd)
	return cmd
}

// customPlugins reads every manifest in pluginManifests and returns its descriptors in order.
func customPlugins(project string, c config.Config) ([]plugins.Descriptor, error) {
	var out []plugins.Descriptor
	for _, path := range c.PluginManifests {
		if !filepath.IsAbs(path) {
			path = filepath.Join(project, path)
		}
		d, err := plugins.Load(path)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func pluginList(project string, out io.Writer) error {
	c, err := config.Load(project)
	if err != nil {
		return err
	}
	custom, err := customPlugins(project, c)
	if err != nil {
		return err
	}
	p := presentationFor(out)
	row := func(id string, perms []string) error {
		state, color := "off", "peach"
		if c.HasPlugin(id) {
			state, color = "on", "mint"
		}
		extra := "-"
		if len(perms) > 0 {
			extra = strings.Join(perms, ", ")
		}
		_, err := fmt.Fprintf(out, "  %s  %-14s %s\n", p.paint(fmt.Sprintf("%-3s", state), color), id, extra)
		return err
	}
	for _, id := range builtinPluginOrder {
		// Permissions the plugin adds by itself, as the manifest builder computes them.
		if err = row(id, config.Config{Plugins: []string{id}}.EffectivePermissions()); err != nil {
			return err
		}
	}
	for _, d := range custom {
		if err = row(d.ID, d.Permissions); err != nil {
			return err
		}
	}
	return nil
}

func pluginEdit(project string, ids []string, add bool, out io.Writer) error {
	c, err := config.Load(project)
	if err != nil {
		return err
	}
	custom, err := customPlugins(project, c)
	if err != nil {
		return err
	}
	valid := append([]string(nil), builtinPluginOrder...)
	for _, d := range custom {
		valid = append(valid, d.ID)
	}
	known := map[string]bool{}
	for _, id := range valid {
		known[id] = true
	}
	var changed []string
	for _, id := range ids {
		if !known[id] {
			return fmt.Errorf("unknown plugin %q; valid plugins: %s", id, strings.Join(valid, ", "))
		}
		if add == c.HasPlugin(id) {
			continue
		}
		if add {
			c.Plugins = append(c.Plugins, id)
		} else {
			var kept []string
			for _, existing := range c.Plugins {
				if existing != id {
					kept = append(kept, existing)
				}
			}
			c.Plugins = kept
		}
		changed = append(changed, id)
	}
	if len(changed) == 0 {
		if add {
			return writeStatus(out, "Unchanged", "the given plugins are already enabled.", "iris")
		}
		return writeStatus(out, "Unchanged", "the given plugins are not enabled; run 'koru plugin list' to see active plugins.", "iris")
	}
	if err = saveConfig(project, c); err != nil {
		return err
	}
	label := "Enabled"
	if !add {
		label = "Disabled"
	}
	return writeStatus(out, label, strings.Join(changed, ", "), "mint")
}

func newConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "config [project]",
		Short:   "Print the resolved configuration as JSON",
		Example: "  koru config ./my-app",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := "."
			if len(args) == 1 {
				project = args[0]
			}
			return printConfig(project, cmd.OutOrStdout())
		},
	}
}

func printConfig(project string, out io.Writer) error {
	c, err := config.Load(project)
	if err != nil {
		return err
	}
	resolve := func(p string) (string, error) {
		if !filepath.IsAbs(p) {
			p = filepath.Join(project, p)
		}
		return filepath.Abs(p)
	}
	webRoot, err := resolve(c.WebDir)
	if err != nil {
		return err
	}
	key, err := resolve(c.SigningKey)
	if err != nil {
		return err
	}
	// The signing key path is reported but the key itself is never read.
	report := struct {
		Config               config.Config `json:"config"`
		EffectivePermissions []string      `json:"effectivePermissions"`
		WebRoot              string        `json:"webRoot"`
		SigningKey           string        `json:"signingKey"`
	}{c, c.EffectivePermissions(), webRoot, key}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
