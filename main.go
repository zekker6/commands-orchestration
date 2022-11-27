package main

import (
	"flag"
	"log"
	"os"

	"commands-orchestration/play"
	"commands-orchestration/updater"

	"gopkg.in/yaml.v2"
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

func readConfig(path string) *play.Play {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	var p play.Play

	err = yaml.Unmarshal(data, &p)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	pl := play.NewPlay(p.Stages, p.Vars)

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

	pl := readConfig(flag.Arg(0))

	pl.Run(*verbose)

	pl.DumpLogs()
	pl.PrintResults()
}
