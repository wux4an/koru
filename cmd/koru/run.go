package main

import (
	"context"
	"fmt"
	"io"
	"koru.dev/koru/internal/adb"
	"koru.dev/koru/internal/app"
	"koru.dev/koru/internal/config"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// launchActivity is the Activity the runtime DEX declares for every app.
const launchActivity = "com.koru.runtime.MainActivity"

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [project]",
		Short: "Build, install and open the app on a device or emulator",
		Long: `Build the APK, install it on a connected Android device or emulator and open it.

koru talks to the adb server that Android Studio or the platform tools run; it
does not start programs itself. Debug builds can be inspected in Chrome at
chrome://inspect.`,
		Example: `  koru run
  koru run ./my-app --device emulator-5554
  koru run --release`,
		Args: cobra.MaximumNArgs(1),
	}
	cmd.Flags().String("device", "", "device serial from `koru devices` (default: the only connected device, or $ANDROID_SERIAL)")
	cmd.Flags().Bool("release", false, "install a release build (debugging disabled)")
	cmd.Flags().Bool("no-cache", false, "disable runtime DEX cache")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		project := "."
		if len(args) == 1 {
			project = args[0]
		}
		serial, _ := cmd.Flags().GetString("device")
		if serial == "" {
			serial = os.Getenv("ANDROID_SERIAL")
		}
		release, _ := cmd.Flags().GetBool("release")
		noCache, _ := cmd.Flags().GetBool("no-cache")
		addr, err := adb.Address()
		if err != nil {
			return err
		}
		return runOnDevice(cmd.Context(), adb.Client{Addr: addr}, project, serial, release, noCache, cmd.OutOrStdout())
	}
	return cmd
}

func runOnDevice(ctx context.Context, client adb.Client, project, serial string, release, noCache bool, out io.Writer) error {
	c, err := config.Load(project)
	if err != nil {
		return err
	}
	// Find the device first so a missing phone fails before a full build.
	devices, err := client.Devices(ctx)
	if err != nil {
		return err
	}
	device, err := adb.Pick(devices, serial)
	if err != nil {
		return err
	}
	start := time.Now()
	b, err := app.Build(project, app.Options{Release: release, NoCache: noCache})
	if err != nil {
		return err
	}
	if err = writeStatus(out, "Built", fmt.Sprintf("%s %s (%d KB) in %s", c.AppID, c.Version, len(b)/1024, time.Since(start).Round(time.Millisecond)), "iris"); err != nil {
		return err
	}
	name := device.Serial
	if device.Model != "" {
		name = device.Model + " (" + device.Serial + ")"
	}
	if err = client.Install(ctx, device.Serial, b); err != nil {
		return err
	}
	if err = writeStatus(out, "Installed", "on "+name, "iris"); err != nil {
		return err
	}
	if err = client.Start(ctx, device.Serial, c.AppID+"/"+launchActivity); err != nil {
		return err
	}
	hint := "inspect it at chrome://inspect"
	if release {
		hint = "release build; debugging is disabled"
	}
	return writeStatus(out, "Opened", fmt.Sprintf("%s on %s; %s", c.Name, name, hint), "mint")
}

func newDevicesCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "devices",
		Short:   "List connected Android devices and emulators",
		Example: "  koru devices",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			addr, err := adb.Address()
			if err != nil {
				return err
			}
			return listDevices(cmd.Context(), adb.Client{Addr: addr}, cmd.OutOrStdout())
		},
	}
}

func listDevices(ctx context.Context, client adb.Client, out io.Writer) error {
	devices, err := client.Devices(ctx)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return writeStatus(out, "No devices.", "Connect a phone with USB debugging enabled, or start an emulator.", "peach")
	}
	p := presentationFor(out)
	for _, d := range devices {
		color := "mint"
		if d.State != "device" {
			color = "peach"
		}
		if _, err = fmt.Fprintf(out, "%-24s %s  %s\n", d.Serial, p.paint(fmt.Sprintf("%-12s", d.State), color), d.Model); err != nil {
			return err
		}
	}
	return nil
}
