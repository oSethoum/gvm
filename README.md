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
gvm init <shell>       # shell integration (powershell|bash|zsh|fish)
gvm doctor             # diagnose setup
gvm version            # print version
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

## Cross-compiling

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gvm-linux-amd64 .
```

## License

MIT
