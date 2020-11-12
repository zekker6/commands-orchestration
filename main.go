package main

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
)

type stage struct {
	Steps []string `yaml:"steps"`
}

type play struct {
	wg        *sync.WaitGroup
	Stages    []stage `yaml:"play"`
	errorChan chan error
}

func NewPlay(stages []stage) *play {
	wg := new(sync.WaitGroup)
	errorChan := make(chan error)

	return &play{
		wg:        wg,
		Stages:    stages,
		errorChan: errorChan,
	}
}

func (p *play) runBackground(command string) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		p.errorChan <- err
	}

	cmd.Wait()
	p.wg.Done()
}

func (p *play) printErrors() {
	for er := range p.errorChan {
		log.Print("Error during execution: ", er)
	}
}

func (p *play) Run() {
	for _, stage := range p.Stages {
		p.wg.Add(len(stage.Steps))
		for _, command := range stage.Steps {
			go p.runBackground(command)
		}
		p.wg.Wait()
	}
}

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	var p play

	err = yaml.Unmarshal(data, &p)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	pl := NewPlay(p.Stages)

	pl.Run()
}
