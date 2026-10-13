#!/usr/bin/env python3
"""Create a reproducible source + Linux executable delivery, without signing keys."""

import argparse
import hashlib
import pathlib
import subprocess
import zipfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=pathlib.Path, required=True)
    args = parser.parse_args()
    root = pathlib.Path(__file__).resolve().parents[1]
    tracked = (
        subprocess.check_output(["git", "ls-files", "-z"], cwd=root)
        .decode()
        .split("\0")
    )
    names = sorted(
        set(filter(None, tracked))
        | {"bin/koru", "artifacts/koru-demo.apk", "artifacts/koru-all-plugins.apk"}
    )
    checksums = []
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(
        args.output, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9
    ) as archive:
        for name in names:
            path = pathlib.PurePosixPath(name)
            if any(
                p in {".git", ".agents", ".koru", "android-demo"} for p in path.parts
            ):
                raise ValueError(f"private directory in delivery: {name}")
            if path.suffix.lower() in {".pem", ".jks", ".p12", ".pyc"}:
                raise ValueError(f"private key or cache in delivery: {name}")
            source = root / name
            if source.is_symlink() or not source.is_file():
                raise ValueError(f"delivery entry is not a regular file: {name}")
            data = source.read_bytes()
            checksums.append(f"{hashlib.sha256(data).hexdigest()}  {name}\n")
            entry = zipfile.ZipInfo(f"koru/{name}", (2026, 10, 13, 0, 0, 0))
            entry.compress_type = zipfile.ZIP_DEFLATED
            entry.create_system = 3
            entry.external_attr = (0o100755 if name == "bin/koru" else 0o100644) << 16
            archive.writestr(entry, data)
        entry = zipfile.ZipInfo("koru/SHA256SUMS", (2026, 10, 13, 0, 0, 0))
        entry.compress_type = zipfile.ZIP_DEFLATED
        entry.external_attr = 0o100644 << 16
        archive.writestr(entry, "".join(checksums))
    with zipfile.ZipFile(args.output) as archive:
        if archive.testzip() is not None:
            raise ValueError("delivery ZIP integrity failure")
    print(f"{args.output}: {len(names)} files, {args.output.stat().st_size} bytes")


if __name__ == "__main__":
    main()
