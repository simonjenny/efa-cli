// Command efa is the Go port of the efa-cli command line tool for
// Electronic Timetable Information (VRS/EFA).
package main

import (
	"os"

	"github.com/simonjenny/efa-cli/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
