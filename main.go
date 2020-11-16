package main

import (
	"commands-orchestration/play"
	"commands-orchestration/updater"
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
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

func main() {
	v := flag.Bool("v", false, "Print version end exit")
	update := flag.Bool("u", false, "Check if newer version is available and self-update")
	flag.Parse()

	if *v {
		printVersion()
		return
	}

	if len(os.Args) < 2 {
		printUsage()
		printVersion()
		return
	}

	if *update {
		updater.DoSelfUpdate(version)
		return
	}

	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	var p play.Play

	err = yaml.Unmarshal(data, &p)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	pl := play.NewPlay(p.Stages, p.Vars)

	pl.Run()
	log.Print("Finished run")

	pl.DumpLogs()
	pl.PrintResults()
}
