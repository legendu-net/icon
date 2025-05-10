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
    cmd = f"""cd {SCRIPT_DIR.parent} && \
        ./icon completion bash > utils/data/bash-it/icon.completion.bash
        """
    sp.run(cmd, shell=True, check=True)


def gen_completion_script_ldc() -> None:
    print("Generate completion script for ldc...")
    completely = 'docker run --rm -it --user $(id -u):$(id -g) --volume "$PWD:/app" dannyben/completely'
    cmd = f"""{completely} preview > completely.bash \
        && mv completely.bash ../utils/data/bash-it/ldc.completion.bash
        """
    sp.run(cmd, shell=True, check=True)


def main() -> None:
    build_project()
    gen_completion_script_icon()
    gen_completion_script_ldc()


if __name__ == "__main__":
    main()
