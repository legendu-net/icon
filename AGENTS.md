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
  `shell`, `virtualization`. A category package normally imports only `cmd/icon` and `cmd/network`
  (plus `utils`), but it may also import another category package when one tool has to install
  another — `cmd/filesystem` imports `cmd/dev` for `dev.InstallJjTools`, because the Yazi plugin
  `Adda0/jjui` needs `jj` and `jjui` at runtime. Such an edge makes the dependency between the two
  packages directional, so keep it one-way to avoid an import cycle.
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

**The configuration of every app lives in the separate `legendu-net/icon-data` repo, not in this
repo.** `icon data` (`cmd/icon/data.go`) clones it into `~/.config/icon-data`; `--config` actions
call `icon.FetchConfigData(false, "")` first, which is a no-op when the clone already exists.

**Never edit `~/.config/icon-data` — it is a pull target, not a working copy.** `icon <app> -c`
reads from it, and `icon data` / `icon data --force` (re)pulls it, the latter renaming the existing
copy to a timestamped backup and re-cloning, so anything written there is silently dropped out of
use. Development on the configuration data happens in the sibling checkout **`../icon-data`**:
extract or edit configuration there, commit and push it, and it reaches `~/.config/icon-data` on
the next `icon data --force`.

Layout: one directory per app, usually named after the subcommand — `~/.config/icon-data/<app>/…`
(e.g. `zellij/config.kdl`, `waveterm/settings.json`, `yazi/keymap.toml`), though a few follow the
app's own spelling instead (`neovim` → `nvim`, `bash_it` → `bash-it`). `user.yaml` holds the shared
user identity (`utils.ReadUserConfig`), so never hard-code a name or email in a command.

Track hand-written configuration only. A file an app generates and rewrites itself (e.g. Yazi's
`package.toml`, a lock file maintained by `ya pkg`) stays machine-local: symlinking it would make
the app write into `~/.config/icon-data`, and the write would be lost on the next
`icon data --force` anyway.

The standard `--config` body is:

```go
icon.FetchConfigData(false, "")
src := "~/.config/icon-data/<app>/<file>"   // fail via log.Fatal if a required src is missing
dst := "~/.config/<app>/<file>"
utils.BackupOrRemove(dst, utils.ShouldBackup(cmd))   // honors --no-backup
utils.CopyOrSymlink(src, dst, utils.GetBoolFlag(cmd, "copy"))   // honors --copy
```

Link the whole app directory (`cmd/shell/zellij.go`) when the app owns all of it, or entry by entry
(`cmd/filesystem/yazi.go`) when the app writes machine-local state into the same directory — e.g.
`~/.config/yazi/plugins`, which `ya pkg` manages and which therefore must not be symlinked into
icon-data. Respect an app's own config-home environment variables (Yazi honors `YAZI_CONFIG_HOME`
and `XDG_CONFIG_HOME`) instead of hard-coding `~/.config/<app>`.

**Supporting a new app therefore takes changes in two repos:** the command here, and its
configuration committed and pushed to `legendu-net/icon-data`, without which `--config` fails on
any other machine.

## Release / CI flow

- The `icon` version is defined in the `version` function in `cmd/icon/version.go` (a hard-coded
  string); bump it there when cutting a release.
- Pushing any non-`main` branch auto-opens a PR to `main` (`create_pr_to_main.yml`); stale branches
  are pruned nightly (`remove_branch.yml`).
- Publishing a GitHub release builds cross-platform binaries (`release.yml`, linux/darwin ×
  amd64/arm64) and dispatches an event to `legendu-net/podman` (`dispatch.yml`).
- `install_icon.sh` is the user-facing installer that downloads the latest released binary.
