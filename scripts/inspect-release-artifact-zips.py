#!/usr/bin/env python3
import argparse
import stat
import sys
import zipfile


def fail(message):
    raise ValueError(message)


def inspect(paths, max_entries, max_expanded_bytes):
    entries = 0
    expanded_bytes = 0
    for path in paths:
        try:
            with zipfile.ZipFile(path) as archive:
                seen = set()
                for entry in archive.infolist():
                    entries += 1
                    if entries > max_entries:
                        fail("entry count exceeds limit")
                    expanded_bytes += entry.file_size
                    if expanded_bytes > max_expanded_bytes:
                        fail("expanded size exceeds limit")

                    name = entry.filename
                    trimmed = name[:-1] if entry.is_dir() else name
                    if not trimmed or "\\" in trimmed or "\0" in trimmed or trimmed.startswith("/"):
                        fail(f"unsafe path: {name!r}")
                    if any(part in ("", ".", "..") for part in trimmed.split("/")):
                        fail(f"unsafe path: {name!r}")
                    if name in seen:
                        fail(f"duplicate path: {name!r}")
                    seen.add(name)

                    kind = stat.S_IFMT(entry.external_attr >> 16)
                    allowed = (0, stat.S_IFDIR) if entry.is_dir() else (0, stat.S_IFREG)
                    if kind not in allowed:
                        fail(f"non-regular entry: {name!r}")
                    if entry.flag_bits & 1:
                        fail(f"encrypted entry: {name!r}")
        except zipfile.BadZipFile as error:
            fail(f"invalid ZIP {path}: {error}")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--max-entries", required=True, type=int)
    parser.add_argument("--max-expanded-bytes", required=True, type=int)
    parser.add_argument("archives", nargs="+")
    args = parser.parse_args()
    if args.max_entries < 1 or args.max_expanded_bytes < 1:
        parser.error("limits must be positive")
    inspect(args.archives, args.max_entries, args.max_expanded_bytes)


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError) as error:
        print(f"release artifact inspection: {error}", file=sys.stderr)
        sys.exit(1)
