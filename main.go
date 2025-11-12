package main

import (
	"flag"
	"log"
	"os"

	"commands-orchestration/play"
	"commands-orchestration/updater"

	"github.com/mattn/go-isatty"
	"gopkg.in/yaml.v3"
)

var (
	version string
	commit  string
	date    string
)

func printVersion() {
	log.Printf("Version: %s, Commit: %s, Date: %s", version, commit, date)
}

func printUsage() {
	log.Print("Specify config path")
	log.Print("Usage: ", os.Args[0], " [config path]")

	flag.Usage()
}

func readConfig(path string, hasTTY bool) *play.Play {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	var p play.Play

	err = yaml.Unmarshal(data, &p)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	pl := play.NewPlay(p.Stages, p.Vars, hasTTY)

	return pl
}

func main() {
	v := flag.Bool("v", false, "Print version end exit")
	update := flag.Bool("u", false, "Check if newer version is available and self-update")
	verbose := flag.Bool("vv", false, "Verbose output")
	flag.Parse()

	if *v {
		printVersion()
		return
	}

	if *update {
		updater.DoSelfUpdate(version)
		return
	}

	if len(os.Args) < 2 {
		printUsage()
		printVersion()
		return
	}

	hasTTY := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	pl := readConfig(flag.Arg(0), hasTTY)

	pl.Run(*verbose)

	pl.DumpLogs()
	pl.PrintResults()
}
