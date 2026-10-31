import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const repo = resolve(import.meta.dir, "../..");
export const releasesURL = "https://github.com/wux4an/koru/releases";

/** The koru release, read from the CLI so the site never invents one. */
export function koruVersion() {
  const source = readFileSync(resolve(repo, "cmd/koru/main.go"), "utf8");
  const match = source.match(
    /^[ \t]*(?:(?:var|const)[ \t]+)?version[ \t]*=[ \t]*"([^"]+)"/m,
  );
  if (!match)
    throw new Error(
      "cmd/koru/main.go has no `version` declaration (var, const or inside a var block).",
    );
  return match[1];
}

/** Display form of the build version, linked to its release page. */
export const versionTag = () => `v${koruVersion()}`;
export const releaseURL = () => `${releasesURL}/tag/${versionTag()}`;
