# gvm — Go Version Manager

A polished, cross-platform Go version manager with a Charmbracelet-powered TUI.

## Features

- Install any Go release from go.dev
- Switch versions instantly (shims + env vars)
- Interactive searchable lists (Charmbracelet/bubbles)
- Per-version GOPATH (isolated `go install` binaries)
- Shared module cache across versions
- Windows / macOS / Linux

## Install

### go install

```sh
go install github.com/osethoum/gvm@latest
```

### Binary (GitHub Releases)

Unix/macOS:
```sh
curl -sSL https://raw.githubusercontent.com/oSethoum/gvm/main/install.sh | sh
```

Windows (PowerShell):
```powershell
irm https://raw.githubusercontent.com/oSethoum/gvm/main/install.ps1 | iex
```

## Usage

```sh
gvm                    # interactive menu
gvm install [version]  # install (interactive if omitted)
gvm use [version]      # switch (interactive if omitted)
gvm list               # searchable list of versions
gvm current            # show active version
gvm available          # browse available releases
gvm uninstall          # remove a version
gvm env                # print env vars
gvm move               # move everything to a new folder (interactive browser)
gvm move D:\gvm        # or pass the destination directly (--yes to skip confirm)
gvm init <shell>       # shell integration (powershell|bash|zsh|fish)
gvm doctor             # diagnose setup
gvm version            # print version
gvm update             # update gvm to the latest release
```

## Shell integration

PowerShell:
```powershell
gvm init powershell | Out-File -Append $PROFILE
. $PROFILE
```

Bash/Zsh:
```sh
gvm init bash >> ~/.bashrc   # or ~/.zshrc
source ~/.bashrc
```

## Moving gvm

`gvm move` relocates the whole gvm tree (versions, per-version GOPATHs, module
cache, download cache, shims, config and the active version). It copies into a
sibling staging folder with a progress bar, verifies the copy, and only then
swaps it into place and regenerates shims, the `current` link, `GVM_ROOT` in
your shell profile, and the Windows user environment (`GOROOT`, `GOPATH`,
`GOBIN`, `GOMODCACHE` and `PATH`). The gvm executable is copied into the new
tree as well and PATH entries are repointed at it.

Interrupting with Ctrl+C cancels the copy and removes the partial staging
folder; the old directory and all env vars stay untouched. Stale staging
folders from an interrupted run are cleaned up automatically next time.

With no argument, `gvm move` opens a folder browser:

- `up`/`down` (or `j`/`k`) move through folders
- `enter`/`right` opens a folder, `left` goes to the parent
- `v` switches to the volumes list on Windows (C:, D:, ...)
- `s` uses the folder currently shown; `q` cancels

For scripts, use `gvm move --dir /path/to/gvm --yes`.

## Directory layout

```
~/.gvm/
  versions/<ver>/     # GOROOT (toolchain)
  gopath/<ver>/       # GOPATH (bin, src, pkg) per version
  modcache/           # shared module cache
  shims/              # go/gofmt wrappers
  active              # active version
```

Projects can live anywhere with a `go.mod` — no need for GOPATH/src.

## License

MIT
