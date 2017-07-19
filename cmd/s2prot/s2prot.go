/*

Package main is a simple CLI app to parse and display information about
a StarCraft II replay passed as a CLI argument.

*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
)

const (
	appName    = "s2prot"
	appVersion = "v1.1.0"
	appAuthor  = "Andras Belicza"
	appHome    = "https://github.com/icza/s2prot"
)

// Flag variables
var (
	version = flag.Bool("version", false, "print version info and exit")

	header      = flag.Bool("header", true, "print replay header")
	details     = flag.Bool("details", false, "print replay details")
	initData    = flag.Bool("initdata", false, "print replay init data")
	attrEvts    = flag.Bool("attrevts", false, "print attributes events")
	metadata    = flag.Bool("metadata", true, "print game metadata")
	gameEvts    = flag.Bool("gameevts", false, "print game events")
	msgEvts     = flag.Bool("msgevts", false, "print message events")
	trackerEvts = flag.Bool("trackerevts", false, "print tracker events")

	indent = flag.Bool("indent", true, "use indentation when formatting output")
)

func main() {
	flag.Parse()

	if *version {
		printVersion()
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	r, err := rep.NewFromFileEvts(args[0], *gameEvts, *msgEvts, *trackerEvts)
	if err != nil {
		fmt.Printf("Failed to parse replay: %v\n", err)
		os.Exit(2)
	}

	// Zero values in replay the user do not wish to see:
	if !*header {
		r.Header.Struct = nil
	}
	if !*details {
		r.Details.Struct = nil
	}
	if !*initData {
		r.InitData.Struct = nil
	}
	if !*attrEvts {
		r.AttrEvts.Struct = nil
	}
	if !*metadata {
		r.Metadata.Struct = nil
	}
	if !*gameEvts {
		r.GameEvts = nil
	}
	if !*msgEvts {
		r.MessageEvts = nil
	}
	if !*trackerEvts {
		r.TrackerEvts = nil
	}

	enc := json.NewEncoder(os.Stdout)
	if *indent {
		enc.SetIndent("", "  ")
	}
	enc.Encode(r)
}

func printVersion() {
	fmt.Println(appName, "version:", appVersion)
	fmt.Println("Parser version:", rep.ParserVersion)
	fmt.Println("Supported replay builds:", s2prot.MinBaseBuild, "..", s2prot.MaxBaseBuild)
	fmt.Println("Platform:", runtime.GOOS, runtime.GOARCH)
	fmt.Println("Built with:", runtime.Version())
	fmt.Println("Author:", appAuthor)
	fmt.Println("Home page:", appHome)
}

func printUsage() {
	fmt.Println("Usage:")
	name := os.Args[0]
	fmt.Printf("\t%s [FLAGS] repfile.SC2Replay\n", name)
	fmt.Println("\tRun with '-h' to see a list of available flags.")
}
