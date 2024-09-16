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

min. PHP 8.2

## Installation

```bash 
wget https://github.com/simonjenny/efa-cli/blob/master/builds/efa?raw=true -O efa
chmod a+x efa && sudo mv efa /usr/local/bin
```

## Demo

[![asciicast](https://asciinema.org/a/yKMgOGa3LiOn1CBIDtN8AqSE4.svg)](https://asciinema.org/a/yKMgOGa3LiOn1CBIDtN8AqSE4)


