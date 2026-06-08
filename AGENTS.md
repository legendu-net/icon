# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Overview

`icon` is a Go CLI (built on Cobra) that installs and—more importantly—configures development
tools, AI/big-data frameworks, IDEs, and shell utilities on Linux and macOS. Each tool is exposed
as a top-level subcommand (e.g. `icon golang -ic`).

On Linux, the focus is on the Debian/Ubuntu series (especially Ubuntu) and the Fedora series of
distributions.

## Commands

- **Build:** `go build` (produces the `icon` binary)
- **Run:** `./icon <tool> [flags]`, e.g. `./icon golang -ic` (install + config Golang)
- **Lint (CI uses golangci-lint v2):**
  - `golangci-lint fmt -d` — check formatting (must produce no diff)
  - `GOFLAGS=-buildvcs=false golangci-lint run` — lint
- **Shell script checks** (for `install_icon.sh`):
  - `shfmt -i 4 -ci -d install_icon.sh` — formatting
  - `shellcheck install_icon.sh` — lint

There are no Go unit tests in this repo; verification is done by building and running the CLI.

## Architecture

- `main.go` → `cmd.Execute()` in `cmd/root.go`. `Execute()` rejects non-darwin/linux OSes, then
  registers every subcommand by calling each `Config<Tool>Cmd(rootCmd)`. **Adding a new tool means
  adding a `Config<Tool>Cmd(rootCmd)` call here** — it is the single registry of all commands.
- `cmd/` is organized by category packages: `ai`, `bigdata`, `dev`, `filesystem`, `icon` (the tool's
  own meta-commands: `data`, `update`, `version`, `completion`), `ide`, `jupyter`, `misc`, `network`,
  `shell`, `virtualization`.
- `utils/` is the shared library all commands build on. Prefer these over raw stdlib calls for
  consistency: `RunCmd`/`Format` (shell exec with `{placeholder}` templating), `GetCommandPrefix`
  (decides whether to prepend `sudo` based on path write-permissions), `Get*Flag`, OS detection
  (`IsLinux`, `IsDebianSeries`, `IsFedoraSeries`, `IsAtomicLinux`, `HostKernelArch`, …), filesystem
  helpers (`fs.go`, `fs_shell.go`), and `DownloadFile`/HTTP helpers in `network.go`.

## Command conventions

Each tool file follows the same pattern (see `cmd/dev/golang.go` as the canonical example):

1. A `Config<Tool>Cmd(rootCmd *cobra.Command)` func defines flags and calls `rootCmd.AddCommand(...)`.
2. A `&cobra.Command{ ... Run: <tool> }` var, where the `Run` handler reads bool flags and branches:
   - `--install` / `-i` → install logic
   - `--config` / `-c` → configuration logic (write dotfiles, symlink, etc.)
   - `--uninstall` / `-u` → uninstall (often a no-op placeholder)
3. Config-writing commands also commonly define `--no-backup` and `--copy` (symlink vs. copy) flags;
   use `utils.ShouldBackup(cmd)` and the `CopyOrSymlink` helpers.

Install commands use `GetCommandPrefix(...)` to compute an empty or `sudo`/`sudo -E` prefix and pass
it into `Format` templates so privilege escalation only happens when the target paths aren't writable.

**Error handling:** terminal errors use `log.Fatal` (standard for this CLI); `main.go` sets
`log.Lshortfile` so failures report their source location.

## Configuration data

The `icon data` command (`cmd/icon/data.go`) clones the separate **`legendu-net/icon-data`** repo
into `~/.config/icon-data`. That external repo holds the dotfile/config templates many `--config`
actions copy or symlink into place; some tools depend on it being fetched first.

## Release / CI flow

- The `icon` version is defined in the `version` function in `cmd/icon/version.go` (a hard-coded
  string); bump it there when cutting a release.
- Pushing any non-`main` branch auto-opens a PR to `main` (`create_pr_to_main.yml`); stale branches
  are pruned nightly (`remove_branch.yml`).
- Publishing a GitHub release builds cross-platform binaries (`release.yml`, linux/darwin ×
  amd64/arm64) and dispatches an event to `legendu-net/podman` (`dispatch.yml`).
- `install_icon.sh` is the user-facing installer that downloads the latest released binary.
