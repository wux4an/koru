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
		p := presentationFor(os.Stderr)
		fmt.Fprintln(os.Stderr, p.paint("koru:", "peach"), err)
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
		Use:   "koru",
		Short: "Build native Android packages from web projects",
		Long:  "Create, preview, run and ship Android apps made with HTML, CSS and JavaScript.\nkoru is a self-contained Go toolchain: no Java, Gradle or Android SDK needed to build.",
		Example: `  koru create my-app        # start from a template
  koru dev                  # preview in a browser with live reload
  koru run                  # install and open on a phone or emulator
  koru logs                 # stream the app's logs from the device
  koru plugin add vibration # enable a plugin in koru.json
  koru bump minor           # raise version and versionCode
  koru build --release --bump --output app-release.apk`,
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
	root.SetUsageTemplate(helpTemplate(presentationFor(out)))
	root.AddCommand(newCreateCommand(), newDevCommand(), newRunCommand(), newLogsCommand(), newDevicesCommand(), newJSCommand(), newVersionCommand(), newDoctorCommand(), newAPKBuildCommand(), newInitCommand(), newKeygenCommand(), newDeepLinksCommand(), newVerifyCommand(), newDexCommand(), newBumpCommand(), newPluginCommand(), newConfigCommand())
	return root
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print the koru version",
		Example: "  koru version",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			p := presentationFor(out)
			_, err := fmt.Fprintf(out, "%s (%s/%s)\n", p.paint(fmt.Sprintf("koru %s", version), "iris"), runtime.GOOS, runtime.GOARCH)
			return err
		},
	}
}

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "doctor [project]",
		Short:   "Check a project's configuration and signing setup",
		Example: "  koru doctor ./my-app",
		Args:    cobra.MaximumNArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error { return doctor(args, cmd.OutOrStdout()) },
	}
}

func newAPKBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "build [project]",
		Short:   "Build and sign an Android APK",
		Example: "  koru build ./my-app --output app.apk\n  koru build ./my-app --release --output app-release.apk",
		Args:    cobra.MaximumNArgs(1),
	}
	cmd.Flags().String("output", "app.apk", "APK output path")
	cmd.Flags().String("keystore", "", "existing PEM signing identity")
	cmd.Flags().Bool("no-cache", false, "disable runtime DEX cache")
	cmd.Flags().Bool("release", false, "disable debug mode and reuse the persistent identity")
	cmd.Flags().Bool("bump", false, "increase versionCode and the patch version in koru.json before building")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		key, _ := cmd.Flags().GetString("keystore")
		noCache, _ := cmd.Flags().GetBool("no-cache")
		release, _ := cmd.Flags().GetBool("release")
		bump, _ := cmd.Flags().GetBool("bump")
		project := "."
		if len(args) == 1 {
			project = args[0]
		}
		if bump {
			if err := bumpVersion(project, "patch", cmd.OutOrStdout()); err != nil {
				return err
			}
		}
		return apkBuild(project, output, key, noCache, release, cmd.OutOrStdout())
	}
	return cmd
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "init [directory]",
		Short:   "Create a minimal project in place (see create for templates)",
		Example: "  koru init ./my-app",
		Args:    cobra.MaximumNArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error { return initProject(args, cmd.OutOrStdout()) },
	}
}

func newKeygenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keygen",
		Short:   "Create a new PEM signing identity",
		Example: "  koru keygen --output signing.pem",
		Args:    cobra.NoArgs,
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
		Use:     "verify <app.apk>",
		Short:   "Verify an APK signature and package structure",
		Example: "  koru verify app.apk",
		Args:    cobra.ExactArgs(1),
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
	cmd := &cobra.Command{
		Use:     "dex",
		Short:   "Build and inspect Dalvik Executable files",
		Example: "  koru dex inspect classes.dex",
	}
	buildCmd := &cobra.Command{
		Use:     "build [ir.json]",
		Short:   "Compile DEX from a JSON intermediate representation",
		Example: "  koru dex build examples/dex/answer.json --output answer.dex\n  koru dex build --activity-experiment --output experiment.dex",
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
		Use:     "inspect <file.dex>",
		Short:   "Inspect a DEX file and print its structure as JSON",
		Example: "  koru dex inspect classes.dex",
		Args:    cobra.ExactArgs(1),
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

func helpTemplate(p presentation) string {
	template := `{{brand}}
{{if .HasParent}}  {{.CommandPath}}
{{end}}{{if .HasParent}}{{.Short}}{{else}}{{.Long}}{{end}}

{{printf "%sUsage:%s\n" "{{iris}}" "{{reset}}"}}  {{if .HasAvailableSubCommands}}{{.CommandPath}} [command]{{if .HasAvailableFlags}} [flags]{{end}}{{else}}{{.UseLine}}{{end}}
{{if .HasAvailableSubCommands}}
{{printf "%sCommands:%s\n" "{{mint}}" "{{reset}}"}}{{range .Commands}}{{if .IsAvailableCommand}}{{printf "  %s%-*s%s  %s\n" "{{mint}}" .NamePadding .Name "{{reset}}" .Short}}{{end}}{{end}}{{end}}{{if .Example}}
{{printf "%sExamples:%s\n" "{{peach}}" "{{reset}}"}}{{.Example}}{{end}}{{if .HasAvailableFlags}}

{{printf "%sFlags:%s\n" "{{iris}}" "{{reset}}"}}{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if not .HasParent}}

Run 'koru <command> --help' for command details.{{else}}

Run '{{.Parent.CommandPath}} --help' to see parent commands.{{end}}
`
	template = strings.ReplaceAll(template, "{{brand}}", p.brandLine())
	template = strings.ReplaceAll(template, "{{iris}}", p.code("iris"))
	template = strings.ReplaceAll(template, "{{mint}}", p.code("mint"))
	template = strings.ReplaceAll(template, "{{peach}}", p.code("peach"))
	reset := ""
	if p.mode != colorOff {
		reset = "\x1b[0m"
	}
	return strings.ReplaceAll(template, "{{reset}}", reset)
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
	return writeStatus(out, "Wrote", fmt.Sprintf("%s (%d bytes); DEX 035 verified successfully.", output, len(b)), "mint")
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
