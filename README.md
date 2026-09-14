# Command Line Tool for Electronic Timetable Information

```bash 
  $ efa 
  
  efa-cli  v2.0

  USAGE:  <command> [options] [arguments]

  departures Create a departure schedule for a specific bus stop.
  messages   Show current information, disruptions and alerts (currently only from the Basler Verkehrs-Betriebe network in german!)
  route      Plan a trip from point A to point B (currently only routes available from the current time and date)
  stopinfo   Show Information for a stop.

```

## Requirements

Go 1.27 or newer (for building from source)

## Installation

Build the static binary:

```bash
go build -o efa ./cmd/efa
```

or use the provided Makefile:

```bash
make build
```

Copy the binary to your PATH:

```bash
sudo mv efa /usr/local/bin
```

## Run with Docker

```bash
docker build -t simonjenny/efa .
docker run -it simonjenny/efa OPTIONS ARGUMENTS
```

## Demo

[![asciicast](https://asciinema.org/a/yKMgOGa3LiOn1CBIDtN8AqSE4.svg)](https://asciinema.org/a/yKMgOGa3LiOn1CBIDtN8AqSE4)

## Commands

```
efa departures <stop> [--limit=<limit>] [--gid] [--json]
efa messages [--json]
efa route <start> <destination> [<time>] [<date>] [<mode>] [--json]
efa stopinfo <stop> [--json]
```

Missing arguments are requested interactively (search, text and select prompts).