# ghstats

`ghstats` is a tiny terminal dashboard that turns `git log` output into weekly contribution cards. It buckets every commit by ISO week, aggregates additions/deletions/commits, and renders lightweight charts using [Charmbracelet Lip Gloss](https://github.com/charmbracelet/lipgloss). Run it inside any Git repository to see a high-level "All Contributors" card plus per-author impact snapshots, sparklines, and filters to zero in on the time range or contributor you care about.

## Highlights

- **Team + individual views** – Always shows a combined "All Contributors" card and optional cards per author sorted by overall impact.
- **Terminal-native charts** – Stacked bar charts and sparklines adapt automatically to your terminal width.
- **Flexible filtering** – Limit by date range, author regex, contributor count, or repository directory.
- **Single static binary** – No services to run; just point the executable at a Git repo and read the report.

## Installation

Prebuilt binaries are produced by the release workflow and published to the GitHub Releases page at [`github.com/boi-gg/ghstats/releases`](https://github.com/boi-gg/ghstats/releases). Each release ships cross-compiled artifacts for:

- **Operating systems:** Linux, macOS (Darwin), and Windows
- **Architectures:** `amd64` and `arm64` (note: Windows/arm64 builds are released without UPX compression for compatibility)

### One-line installer script

Pick the command that matches your shell/OS and it will download the latest release, drop it into a user-writable directory, and put executable bits in place.

**macOS / Linux (bash, zsh, sh, fish):**

```sh
curl -fsSL https://raw.githubusercontent.com/boi-gg/ghstats/main/scripts/install.sh | bash
```

**Windows PowerShell / pwsh:**

```powershell
irm https://raw.githubusercontent.com/boi-gg/ghstats/main/scripts/install.ps1 | iex
```

The POSIX installer defaults to `~/.local/bin`; override with `INSTALL_DIR=/custom/path` or `--dir /custom/path`. The PowerShell installer targets `%LOCALAPPDATA%\Programs\ghstats` and accepts `-InstallDir C:\tools` if you prefer a different destination.

### Option 1: Download a release binary (recommended)

1. Visit the [latest release](https://github.com/boi-gg/ghstats/releases/latest).
2. Download the file named `ghstats-<os>-<arch>` (or `.exe` on Windows) that matches your platform.
3. Make it executable and place it on your `PATH`:

   ```sh
   chmod +x ghstats-linux-amd64
   mv ghstats-linux-amd64 ~/.local/bin/ghstats
   ```

4. Run `ghstats --help` to verify the installation.

### Option 2: `go install`

If you already have Go ≥ 1.24 installed:

```sh
go install github.com/boi-gg/ghstats@latest
```

This drops `ghstats` into your `GOBIN` (usually `$GOPATH/bin`).

### Option 3: Build from source

```sh
git clone https://github.com/boi-gg/ghstats.git
cd ghstats
mise install  # ensures the repo's pinned Go toolchain
mise run go build -o ghstats
```

`mise` is optional; you can also use your system Go toolchain as long as it matches the version in `go.mod` (currently Go 1.24.1).

## Usage

Run `ghstats` inside (or pointed at) a Git repository. By default it scans the entire history and prints the team card followed by one card per contributor.

```sh
ghstats -dir ~/Code/project
```

Filter by date range and author, and show only the aggregate card:

```sh
ghstats -since 2024-01-01 -until 2024-06-30 -author "Alice|Bob" -short
```

### CLI flags

| Flag      | Description                                                              |
| --------- | ------------------------------------------------------------------------ |
| `-since`  | Only include commits on or after this date (`YYYY-MM-DD`).               |
| `-until`  | Only include commits on or before this date (`YYYY-MM-DD`).              |
| `-author` | Regular expression to match author names.                                |
| `-limit`  | Limit the number of contributor cards shown (sorted by impact).          |
| `-dir`    | Path to the Git repository (defaults to `.` or the positional argument). |
| `-short`  | Only display the aggregate "All Contributors" card.                      |

Example output (trimmed for brevity):

```text
All Contributors
Commits: 128   +5421 / -2310
2025-W10 | ███████████████████████████████████ (333)
2025-W11 | ████████████████████ (211)
...

Alice Example
Commits: 57   +2100 / -900   Impact: 3057
2025-W10 | ██████████████... (120)
Summary: ▁▃▄▆█▇
```

## Release workflow

The repository ships with `.github/workflows/release.yml`, which automates builds and releases:

- Triggers on pushes to `develop`, pull requests, manual dispatches, and tags starting with `v` (e.g., `v1.2.0`).
- Uses [`jdx/mise-action`](https://github.com/jdx/mise-action) to provision the pinned Go toolchain.
- Cross-compiles the binary matrix (`GOOS={linux,windows,darwin}` × `GOARCH={amd64,arm64}`) with stripped symbols and embeds the Git tag via `-ldflags`.
- Compresses artifacts with `upx --best --lzma` (skipping Windows/arm64 where UPX is not supported) to keep downloads tiny.
- Publishes the binaries as release assets using `softprops/action-gh-release` whenever a tag push occurs.

Because releases are driven by tags, creating a new version is as simple as:

```sh
git tag v1.0.0
git push origin v1.0.0
```

Once the workflow completes, the tagged binaries will appear on the Releases page.

## Development tips

- `go run ./ghstats.go` lets you iterate without building.
- Use `ghstats -short` while developing to skip author cards and focus on the aggregate output.
- If you add new CLI flags, remember to document them here and update any automation relying on default output.

---

Happy graphing! If you build something neat on top of `ghstats`, open an issue or pull request at [`github.com/boi-gg/ghstats`](https://github.com/boi-gg/ghstats).

---

```txt
$ ghstats -short
╭───────────────────────────────────────────────────────────────────────────────╮
│                                                                               │
│ All Contributors                                                              │
│ Commits: 250   +102376 / -45485                                               │
│                                                                               │
│ 2025-W34 | ██████████████████████████████████████████ (17674)                 │
│ 2025-W35 | █████████████████████████████████████████ (17420)                  │
│ 2025-W36 | ██████████████████████████████ (12749)                             │
│ 2025-W37 | ██████████████████████████████████████████████████████████ (24662) │
│ 2025-W38 | ████████████████████████████ (11943)                               │
│ 2025-W39 | ██████████████████████████████████ (14393)                         │
│ 2025-W40 | ███ (888)                                                          │
│ 2025-W41 | ██████████ (4222)                                                  │
│ 2025-W42 | █████████████████████████████████████████ (17350)                  │
│ 2025-W43 | █████████████████████ (8732)                                       │
│ 2025-W44 | █████████████████████ (8903)                                       │
│ 2025-W45 | ██████████████ (6120)                                              │
│ 2025-W46 | ███████ (3055)                                                     │
│                                                                               │
│                                                                               │
╰───────────────────────────────────────────────────────────────────────────────╯
```
