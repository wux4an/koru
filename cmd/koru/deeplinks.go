package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/signing"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newDeepLinksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deeplinks [project]",
		Short: "Print the assetlinks.json that verifies the https deep links",
		Long: "Print the Digital Asset Links file for the https entries of deepLinks in koru.json.\n" +
			"Publish it at https://<host>/.well-known/assetlinks.json so Android opens those links in the app without asking.",
		Example: "  koru deeplinks ./my-app\n  koru deeplinks ./my-app --json > assetlinks.json",
		Args:    cobra.MaximumNArgs(1),
	}
	cmd.Flags().String("keystore", "", "existing PEM signing identity (default: signingKey from koru.json)")
	cmd.Flags().Bool("json", false, "print only the JSON, ready to redirect to a file")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		project := "."
		if len(args) == 1 {
			project = args[0]
		}
		key, _ := cmd.Flags().GetString("keystore")
		only, _ := cmd.Flags().GetBool("json")
		return deepLinks(project, key, only, cmd.OutOrStdout())
	}
	return cmd
}

type assetTarget struct {
	Namespace    string   `json:"namespace"`
	PackageName  string   `json:"package_name"`
	Fingerprints []string `json:"sha256_cert_fingerprints"`
}
type assetStatement struct {
	Relation []string    `json:"relation"`
	Target   assetTarget `json:"target"`
}

// certFingerprint formats the SHA-256 of a DER certificate as AA:BB:..., the
// form Android and Digital Asset Links expect.
func certFingerprint(der []byte) string {
	sum := sha256.Sum256(der)
	parts := make([]string, len(sum))
	for i, b := range sum {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(parts, ":")
}

func deepLinks(project, key string, jsonOnly bool, out io.Writer) error {
	c, err := config.Load(project)
	if err != nil {
		return err
	}
	hosts := c.HTTPSHosts()
	if len(hosts) == 0 {
		return errors.New(`no https deepLinks in koru.json: add {"scheme":"https","host":"example.com"} to "deepLinks" (custom schemes such as "myapp" need no assetlinks.json)`)
	}
	keyPath := c.SigningKey
	if key != "" {
		keyPath = key
	}
	if !filepath.IsAbs(keyPath) {
		keyPath = filepath.Join(project, keyPath)
	}
	id, err := signing.LoadIdentity(keyPath)
	if err != nil {
		return fmt.Errorf("load signing identity %s: %w", keyPath, err)
	}
	statements := []assetStatement{{
		Relation: []string{"delegate_permission/common.handle_all_urls"},
		Target:   assetTarget{Namespace: "android_app", PackageName: c.AppID, Fingerprints: []string{certFingerprint(id.Certificate)}},
	}}
	b, err := json.MarshalIndent(statements, "", "  ")
	if err != nil {
		return err
	}
	if jsonOnly {
		_, err = fmt.Fprintf(out, "%s\n", b)
		return err
	}
	if err = writeStatus(out, "Publish this file at", "https://<host>/.well-known/assetlinks.json for each of: "+strings.Join(hosts, ", "), "iris"); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "\n%s\n\nServe it over HTTPS with status 200, content type application/json and no redirects.\n"+
		"Wildcard hosts (*.example.com) need the file on every subdomain that should open the app.\n"+
		"The fingerprint is the SHA-256 of the certificate in %s: keep that key, a different one stops verification.\n"+
		"Android checks the file at install time (autoVerify); until it passes, links may show a chooser.\n"+
		"The link opens the app on its initial page; the URL is not delivered to JavaScript yet.\n", b, c.SigningKey)
	return err
}
