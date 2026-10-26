package main

import (
	"bufio"
	"fmt"
	"io"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/signing"
	"koru.dev/koru/internal/templates"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	var ids []string
	for _, t := range templates.List {
		ids = append(ids, t.ID)
	}
	cmd := &cobra.Command{
		Use:   "create [directory]",
		Short: "Create a new app from a template",
		Long:  "Create a new app from a template, with its configuration and a persistent signing identity. Missing details are asked for in an interactive terminal.",
		Example: `  koru create my-app
  koru create my-app --template react --name "My App" --app-id com.example.myapp`,
		Args: cobra.MaximumNArgs(1),
	}
	cmd.Flags().StringP("template", "t", "", "starter template: "+strings.Join(ids, ", ")+" (default basic)")
	cmd.Flags().String("name", "", "app name shown on the launcher (default from the directory)")
	cmd.Flags().String("app-id", "", "Android application id, such as com.example.myapp")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		o := createOptions{}
		if len(args) == 1 {
			o.dir = args[0]
		}
		o.template, _ = cmd.Flags().GetString("template")
		o.name, _ = cmd.Flags().GetString("name")
		o.appID, _ = cmd.Flags().GetString("app-id")
		return createProject(o, cmd.InOrStdin(), cmd.OutOrStdout(), isTerminal(cmd.InOrStdin()))
	}
	return cmd
}

type createOptions struct{ dir, template, name, appID string }

func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// createProject fills missing options with prompts when interactive, or with
// defaults derived from the directory otherwise.
func createProject(o createOptions, in io.Reader, out io.Writer, interactive bool) error {
	p := presentationFor(out)
	reader := bufio.NewReader(in)
	ask := func(question, fallback string) (string, error) {
		if !interactive {
			return fallback, nil
		}
		if _, err := fmt.Fprintf(out, "%s %s ", p.paint("?", "iris"), question); err != nil {
			return "", err
		}
		if fallback != "" {
			fmt.Fprintf(out, "(%s) ", fallback)
		}
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}
		if line = strings.TrimSpace(line); line == "" {
			return fallback, nil
		}
		return line, nil
	}
	var err error
	if o.dir == "" {
		if !interactive {
			return fmt.Errorf("give a directory, for example: koru create my-app")
		}
		if o.dir, err = ask("Project directory:", "my-app"); err != nil {
			return err
		}
	}
	if o.name == "" {
		if o.name, err = ask("App name:", templates.DisplayName(o.dir)); err != nil {
			return err
		}
	}
	if o.appID == "" {
		if o.appID, err = ask("Application id:", templates.AppID(o.name)); err != nil {
			return err
		}
	}
	if o.template == "" {
		if interactive {
			for i, t := range templates.List {
				fmt.Fprintf(out, "  %d. %s  %s\n", i+1, p.paint(t.ID, "mint"), t.Description)
			}
		}
		answer, err := ask("Template:", "basic")
		if err != nil {
			return err
		}
		o.template = answer
		for i, t := range templates.List {
			if answer == fmt.Sprint(i+1) {
				o.template = t.ID
			}
		}
	}
	c, err := templates.Write(o.dir, templates.Options{Template: o.template, Name: o.name, AppID: o.appID})
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Join(o.dir, ".koru"), 0700); err != nil {
		return err
	}
	if _, err = signing.CreateIdentity(filepath.Join(o.dir, c.SigningKey)); err != nil {
		return err
	}
	return printNextSteps(out, o.dir, o.template, c)
}

func printNextSteps(out io.Writer, dir, template string, c config.Config) error {
	p := presentationFor(out)
	if err := writeStatus(out, "Created", fmt.Sprintf("%s (%s, %s template)", dir, c.AppID, template), "mint"); err != nil {
		return err
	}
	steps := []string{"cd " + dir}
	if t, err := templates.Find(template); err == nil && t.Vendor != "" {
		steps = append(steps, "npm install", "npm run dev        # browser preview with hot reload", "npm run android    # install and open on a device")
	} else {
		steps = append(steps, "koru dev           # browser preview with live reload", "koru run           # install and open on a device")
	}
	fmt.Fprintf(out, "\n%s\n", p.paint("Next steps:", "iris"))
	for _, s := range steps {
		fmt.Fprintf(out, "  %s\n", s)
	}
	_, err := fmt.Fprintf(out, "\nBack up %s: Android only accepts updates signed with the same key.\n", filepath.Join(dir, c.SigningKey))
	return err
}
