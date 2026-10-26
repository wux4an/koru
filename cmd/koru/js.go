package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"koru.dev/koru/packages/javascript"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newJSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "js [directory]",
		Short: "Copy the JavaScript module and its types for bundlers",
		Long: `Write core.js, core.mjs and core.d.mts into a source folder so bundlers such as
Vite can import the koru API. Run it again with --force after upgrading koru.`,
		Example: `  koru js src/koru
  koru js src/koru --force`,
		Args: cobra.MaximumNArgs(1),
	}
	cmd.Flags().Bool("force", false, "replace existing module files")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		dir := filepath.Join("src", "koru")
		if len(args) == 1 {
			dir = args[0]
		}
		force, _ := cmd.Flags().GetBool("force")
		return writeJSModule(dir, force, cmd.OutOrStdout())
	}
	return cmd
}

func writeJSModule(dir string, force bool, out io.Writer) error {
	files := []struct {
		name string
		data []byte
	}{{"core.js", javascript.Core}, {"core.mjs", javascript.Module}, {"core.d.mts", javascript.Types}}
	if !force {
		for _, f := range files {
			if _, err := os.Lstat(filepath.Join(dir, f.name)); err == nil {
				return fmt.Errorf("%s already exists; use --force to replace the module files", filepath.Join(dir, f.name))
			} else if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	for _, f := range files {
		if err := atomicWrite(filepath.Join(dir, f.name), f.data); err != nil {
			return err
		}
	}
	return writeStatus(out, "Wrote", fmt.Sprintf("%s: import { device, storage } from \"./%s/core.mjs\"", dir, filepath.Base(dir)), "mint")
}
