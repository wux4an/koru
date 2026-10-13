// koru compiles and inspects DEX without executing any external build tools.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"koru.dev/koru/internal/app"
	"koru.dev/koru/pkg/dex"
	runtimedex "koru.dev/koru/runtime/dex"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const version = "0.2.0-dev"
const maxInputSize = 64 << 20

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "koru:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	root := newRootCommand(out, io.Discard)
	root.SetArgs(args)
	return root.Execute()
}

func newRootCommand(out, errOut io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "koru",
		Short:         "Build native Android packages from web projects",
		Long:          "Koru turns static web projects into signed Android APKs with a self-contained Go toolchain.",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	root.SetOut(out)
	root.SetErr(errOut)
	root.SetHelpFunc(func(command *cobra.Command, args []string) {
		command.Print(command.UsageString())
	})
	root.SetUsageTemplate(helpTemplate(colorEnabled(out)))
	root.AddCommand(newVersionCommand(), newDoctorCommand(), newAPKBuildCommand(), newInitCommand(), newKeygenCommand(), newVerifyCommand(), newDexCommand())
	return root
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the koru version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "koru %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
			return err
		},
	}
}

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor [project]",
		Short: "Check a project's configuration and signing setup",
		Args:  cobra.MaximumNArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error { return doctor(args, cmd.OutOrStdout()) },
	}
}

func newAPKBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [project]",
		Short: "Build and sign an Android APK",
		Args:  cobra.MaximumNArgs(1),
	}
	cmd.Flags().String("output", "app.apk", "APK output path")
	cmd.Flags().String("keystore", "", "existing PEM signing identity")
	cmd.Flags().Bool("no-cache", false, "disable runtime DEX cache")
	cmd.Flags().Bool("release", false, "disable debug mode and reuse the persistent identity")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		key, _ := cmd.Flags().GetString("keystore")
		noCache, _ := cmd.Flags().GetBool("no-cache")
		release, _ := cmd.Flags().GetBool("release")
		project := "."
		if len(args) == 1 {
			project = args[0]
		}
		return apkBuild(project, output, key, noCache, release, cmd.OutOrStdout())
	}
	return cmd
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init [directory]",
		Short: "Create a web project with a persistent signing identity",
		Args:  cobra.MaximumNArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error { return initProject(args, cmd.OutOrStdout()) },
	}
}

func newKeygenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Create a new PEM signing identity",
		Args:  cobra.NoArgs,
	}
	cmd.Flags().StringP("output", "o", "", "new identity path (required)")
	_ = cmd.MarkFlagRequired("output")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		path, _ := cmd.Flags().GetString("output")
		return keygen(path, cmd.OutOrStdout())
	}
	return cmd
}

func newVerifyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <app.apk>",
		Short: "Verify an APK signature and package structure",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := readLimit(args[0], 320<<20)
			if err != nil {
				return err
			}
			report, err := app.Verify(b)
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
		},
	}
}

func newDexCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "dex", Short: "Build and inspect Dalvik Executable files"}
	buildCmd := &cobra.Command{
		Use:   "build [ir.json]",
		Short: "Compile DEX from a JSON intermediate representation",
		Args: func(cmd *cobra.Command, args []string) error {
			experiment, _ := cmd.Flags().GetBool("activity-experiment")
			if experiment && len(args) != 0 {
				return errors.New("the Activity experiment takes no IR input")
			}
			if !experiment && len(args) != 1 {
				return errors.New("provide one IR JSON file, or use --activity-experiment")
			}
			return nil
		},
	}
	buildCmd.Flags().String("output", "classes.dex", "DEX output path")
	buildCmd.Flags().Bool("activity-experiment", false, "emit the experimental WebView Activity")
	buildCmd.RunE = func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		experiment, _ := cmd.Flags().GetBool("activity-experiment")
		source := ""
		if len(args) == 1 {
			source = args[0]
		}
		return build(source, output, experiment, cmd.OutOrStdout())
	}
	inspectCmd := &cobra.Command{
		Use:   "inspect <file.dex>",
		Short: "Inspect a DEX file and print its structure as JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := readBounded(args[0])
			if err != nil {
				return err
			}
			file, err := dex.Inspect(b)
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(file)
		},
	}
	cmd.AddCommand(buildCmd, inspectCmd)
	return cmd
}

func helpTemplate(color bool) string {
	cyan, mint, lavender, green, reset := "", "", "", "", ""
	if color {
		cyan, mint, lavender, green, reset = "\x1b[36m", "\x1b[92m", "\x1b[95m", "\x1b[32m", "\x1b[0m"
	}
	template := `{{printf "%sUsage%s:\n" "{{cyan}}" "{{reset}}"}}
  {{.UseLine}}
{{if .Long}}
{{.Long}}
{{end}}{{if .HasAvailableSubCommands}}
{{printf "\n%sCommands%s:\n" "{{mint}}" "{{reset}}"}}{{range .Commands}}{{if .IsAvailableCommand}}{{printf "%s%s%s" "{{green}}" (rpad .Name .NamePadding) "{{reset}}"}} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableFlags}}
{{printf "\n%sFlags%s:\n" "{{lavender}}" "{{reset}}"}}{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .HasParent}}
{{printf "\nRun '%s --help' for command details.\n" .CommandPath}}{{end}}`
	template = strings.ReplaceAll(template, "{{cyan}}", cyan)
	template = strings.ReplaceAll(template, "{{mint}}", mint)
	template = strings.ReplaceAll(template, "{{lavender}}", lavender)
	template = strings.ReplaceAll(template, "{{green}}", green)
	return strings.ReplaceAll(template, "{{reset}}", reset)
}

func colorEnabled(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok || os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func readBounded(path string) ([]byte, error) { return readLimit(path, maxInputSize) }
func readLimit(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !s.Mode().IsRegular() {
		return nil, errors.New("input must be a regular file")
	}
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("input exceeds CLI %d byte limit", limit)
	}
	return b, nil
}

func build(source, output string, experiment bool, out io.Writer) error {
	var p dex.Program
	if experiment {
		if source != "" {
			return errors.New("the Activity experiment takes no IR input")
		}
		p = runtimedex.ActivityExperiment()
	} else {
		if source == "" {
			return errors.New("provide one IR JSON file, or use --activity-experiment")
		}
		b, err := readBounded(source)
		if err != nil {
			return err
		}
		dec := json.NewDecoder(strings.NewReader(string(b)))
		dec.DisallowUnknownFields()
		if err = dec.Decode(&p); err != nil {
			return err
		}
		var extra any
		if err = dec.Decode(&extra); err != io.EOF {
			return errors.New("trailing JSON input")
		}
	}
	if source != "" {
		a, err := filepath.Abs(source)
		if err != nil {
			return err
		}
		b, err := filepath.Abs(output)
		if err != nil {
			return err
		}
		if a == b {
			return errors.New("output must differ from input")
		}
		sourceInfo, err := os.Stat(source)
		if err != nil {
			return err
		}
		if target, err := os.Stat(output); err == nil && os.SameFile(sourceInfo, target) {
			return errors.New("output aliases input")
		}
	}
	b, err := dex.Build(p)
	if err != nil {
		return err
	}
	if err = atomicWrite(output, b); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Wrote %s (%d bytes); DEX 035 verified successfully.\n", output, len(b))
	return err
}

// Write the complete validated file before replacing an existing output.
func atomicWrite(path string, b []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".koru-dex-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(b); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Chmod(0644); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
