# GEMINI.md

## Project Overview

`icon` is a comprehensive Go-based CLI tool designed to automate the installation and more important configuration
of a wide range of development environments, AI tools, big data frameworks, IDEs, and shell utilities.
It provides a unified interface to manage system setup on Linux and macOS,
leveraging the Cobra library for a structured command-line experience.

- **Primary Technologies:** Go, Cobra, Shell scripting.
- **Architecture:** The project follows a modular design
  where each tool or category of tools is implemented as a subcommand within the `cmd/` directory.
  Utility functions for common system tasks are centralized in the `utils/` package.

## Building and Running

The project can be built and run using standard Go commands.

- **Build:**
  ```bash
  go build
  ```
- **Run:**
  ```bash
  ./icon tool [flags]
  ```
  Example: Install and configure Golang
  ```bash
  ./icon golang -ic
  ```
- **Installation:**
  The project includes an `install_icon.sh` script
  to download and install the latest pre-built binary from GitHub.

## Project Structure

- `main.go`: The entry point of the application.
- `cmd/`: Contains the implementation of various subcommands grouped by category:
  - `ai/`: AI-related tools (e.g., PyTorch).
  - `bigdata/`: Big data tools (e.g., Spark, ArrowDB).
  - `dev/`: Development languages and tools (e.g., Go, Rust, Git, Homebrew).
  - `filesystem/`: Filesystem utilities (e.g., Dropbox, ripgrep).
  - `ide/`: IDEs and editors (e.g., VS Code, Neovim, Helix).
  - `jupyter/`: Jupyter-related tools and extensions.
  - `network/`: Networking tools and SSH configuration.
  - `shell/`: Shell environments and terminal emulators (e.g., Fish, Nushell, Alacritty).
  - `virtualization/`: Virtualization and containerization (e.g., Docker, KVM).
- `utils/`: Core utility library for:
  - File and directory operations (`fs.go`, `fs_shell.go`).
  - Command execution (`os.go`).
  - Network and download helpers (`network.go`).
  - OS-specific logic and path normalization.

## Development Conventions

- **Adding New Tools:**
  1. Create or update a package in `cmd/` for the tool's category.
  1. Implement a `Config<Tool>Cmd` function to define the subcommand and its flags.
  1. Register the new command in `cmd/root.go`.
- **Command Flags:** Most commands follow a standard set of flags:
  - `--install`, `-i`: Trigger installation logic.
  - `--config`, `-c`: Trigger configuration logic (e.g., writing dotfiles, symlinking).
  - `--uninstall`, `-u`: Trigger uninstallation logic (often a placeholder).
- **Utility Usage:** Prefer using functions in the `utils/` package for system interactions
  to maintain consistency and ensure proper error handling and path normalization.
- **Error Handling:** The current pattern is to use `log.Fatal` for terminal errors,
  which is standard for this CLI tool.
