#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.14"
# ///
"""Print the icon project version parsed from the `version` function in cmd/icon/version.go.

Run with: ./version.py   (or)   uv run version.py
"""

import re
import sys
from pathlib import Path

VERSION_GO = Path(__file__).resolve().parent / "cmd" / "icon" / "version.go"


def parse_version(path: Path = VERSION_GO) -> str:
    """Parse the version string printed by the `version` function in version.go."""
    text = path.read_text(encoding="utf-8")
    # Match a semver-shaped literal so the parse stays correct even if other
    # fmt.Println calls are added to the file.
    match = re.search(r'fmt\.Println\(\s*"(v?\d+\.\d+\.\d+[^"]*)"\s*\)', text)
    if not match:
        raise ValueError(f"Could not parse version from {path}.")
    return match.group(1)


def main() -> int:
    try:
        print(parse_version())
    except Exception as e:
        print(str(e), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
