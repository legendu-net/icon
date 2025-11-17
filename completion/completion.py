#!/usr/bin/env -S uv run

# /// script
# requires-python = ">=3.13"
# dependencies = []
# ///
from pathlib import Path
import subprocess as sp

SCRIPT_DIR = Path(__file__).parent


def build_project() -> None:
    print("Buidling project...")
    cmd = f"cd {SCRIPT_DIR.parent} && go build"
    sp.run(cmd, shell=True, check=True)


def gen_completion_script_icon() -> None:
    print("Generate completion script for icon...")
    cmd = f"""cd {SCRIPT_DIR.parent} \
        && ./icon completion bash > utils/data/bash-it/completion/icon.completion.bash \
        && ./icon completion fish > utils/data/fish/completions/icon.fish
        """
    sp.run(cmd, shell=True, check=True)


def gen_completion_script_ldc() -> None:
    print("Generate completion script for ldc...")
    cmd = f"""cd {SCRIPT_DIR.parent} \
        && crazy-complete --input-type=yaml bash completion/ldc.yaml \
            > utils/data/bash-it/completion/ldc.completion.bash \
        && crazy-complete --input-type=yaml fish completion/ldc.yaml \
            > utils/data/fish/completions/ldc.fish
        """
    sp.run(cmd, shell=True, check=True)


def main() -> None:
    build_project()
    gen_completion_script_icon()
    gen_completion_script_ldc()


if __name__ == "__main__":
    main()
