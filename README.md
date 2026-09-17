# Command Line Tool for Electronic Timetable Information

```bash 
  $ efa 
  
  efa-cli  v3.0

  VERWENDUNG:  <command> [options] [arguments]

  departures Erstelle einen Abfahrtsplan für eine bestimmte Haltestelle.
  mcp        Starte einen MCP-Server, der departures, messages, route und stopinfo als Tools bereitstellt.
  messages   Zeige aktuelle Informationen, Störungen und Meldungen (aktuell nur aus dem Netz der Basler Verkehrs-Betriebe!)
  route      Plane eine Reise von Punkt A nach Punkt B
  stopinfo   Zeige Informationen zu einer Haltestelle.

```

The output language defaults to German and follows the first of `LANGUAGE`, `LC_ALL`, `LC_MESSAGES` or `LANG` that resolves to a supported locale (`C`/`POSIX` and unset variables are ignored); pass `--lang=de` or `--lang=en` to override it explicitly.

## Global Options

| Option | Description |
| --- | --- |
| `-h`, `--help` | Display help for the given command (or for `list` if no command is given) |
| `-q`, `--quiet` | Do not output any message |
| `-V`, `--version` | Display this application version |
| `--ansi` / `--no-ansi` | Force (or disable) ANSI output |
| `-n`, `--no-interaction` | Do not ask any interactive question |
| `-v`, `-vv`, `-vvv` | Increase the verbosity of messages |
| `--lang=<lang>` | Language for output (`de`, `en`) [default: auto] |

`efa help <command>` shows the full help for a command, `efa list` prints the command overview shown above (accepts `--raw`, `--format=<fmt>`, `--short`).

## Requirements

Go 1.27 or newer (for building from source)

## Installation

### Prebuilt binary

Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/simonjenny/efa-cli/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/simonjenny/efa-cli/main/install.ps1 | iex
```

Both scripts fetch the correct binary for your platform from the
[latest release](https://github.com/simonjenny/efa-cli/releases/latest) and install it
(`/usr/local/bin/efa`, or `%LOCALAPPDATA%\efa\efa.exe` on Windows, added to your `PATH`).
Windows SmartScreen may still warn on first run ("Windows protected your PC") since the
binary isn't code-signed — click "More info" → "Run anyway".

Prefer not to run a script? Download the matching binary directly from the
[latest release](https://github.com/simonjenny/efa-cli/releases/latest) instead.

### Build from source

Build the static binary:

```bash
go build -o efa ./cmd/efa
```

or use the provided Makefile:

```bash
make build         # build ./builds/efa
make build-static  # alias for build
make release       # copy builds/efa to release/efa
make test          # go test ./...
make vet           # go vet ./...
make fmt           # list files needing gofmt
make clean         # remove builds/, release/, static/
```

Copy the binary to your PATH:

```bash
sudo mv efa /usr/local/bin
```

## Run with Docker

A prebuilt image is published to the GitHub Container Registry with every [release](https://github.com/simonjenny/efa-cli/releases), tagged with the release version and `latest`:

```bash
docker run -it ghcr.io/simonjenny/efa-cli:latest OPTIONS ARGUMENTS
```

Or build it yourself:

```bash
docker build -t simonjenny/efa .
docker run -it simonjenny/efa OPTIONS ARGUMENTS
```

The image has no default command, so running it without arguments shows the same command overview as running `efa` without arguments.

## Demo (v2.0)

[![asciicast](https://asciinema.org/a/yKMgOGa3LiOn1CBIDtN8AqSE4.svg)](https://asciinema.org/a/yKMgOGa3LiOn1CBIDtN8AqSE4)

## Commands

```
efa departures <stop> [--limit[=<limit>]] [--gid] [--json]
efa messages [--json]
efa route <start> <destination> [<time>] [<date>] [<mode>] [--json]
efa stopinfo <stop> [--json]
```

Missing arguments are requested interactively (search, text and select prompts).

- `--limit`: maximum number of departures to return, default 10. Also applies if `--limit` is given without a value.
- `--gid`: treat `<stop>` as an EFA stop GID (a stable stop identifier) instead of a free-text name, skipping the interactive stop search.
- `<mode>` for `route`: `Departure` or `Arrival`.

`--json` outputs the upstream EFA API's objects largely as-is (e.g. `departures --json` returns the `stopEvents` array from the EFA `rapidJSON` response, with fields like `location.name`, `departureTimePlanned`/`departureTimeEstimated` and `transportation.name`/`.number`). `departures --json` additionally adds a computed `realtime` field: minutes from now until the (estimated or planned) departure.

## MCP Server

Expose `departures`, `messages`, `route` and `stopinfo` as tools over Streamable HTTP for use with MCP-compatible clients (e.g. Claude):

```bash
efa mcp [--ip=<ip>] [--port=<port>]
```

Defaults to `127.0.0.1:8090`.

Example `compose.yaml` using the prebuilt GHCR image:

```yaml
services:
  efa-mcp:
    image: ghcr.io/simonjenny/efa-cli:latest
    command: ["mcp", "--ip=0.0.0.0", "--port=8090"]
    ports:
      - "8090:8090"
```

Or build it locally instead of pulling from GHCR:

```yaml
services:
  efa-mcp:
    build: .
    command: ["mcp", "--ip=0.0.0.0", "--port=8090"]
    ports:
      - "8090:8090"
```

Bind to `0.0.0.0` so the server is reachable from outside the container.