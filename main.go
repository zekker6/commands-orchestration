package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/martinlindhe/notify"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"time"
)

var AvailableColors = []color.Attribute{
	color.FgGreen,
	color.FgYellow,
	color.FgBlue,
	color.FgMagenta,
	color.FgCyan,
}

func randomColor() color.Attribute {
	return AvailableColors[rand.Intn(len(AvailableColors))]

}

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

	handleErr := func(err error) bool {
		if err != nil {
			p.errorChan <- err
			p.wg.Done()

			notify.Alert("co", "Command failed", fmt.Sprintf("%s exited with %+v", command, err), "")
			return false
		}

		return true
	}

	stdout, e := cmd.StdoutPipe()
	if !handleErr(e) {
		return
	}

	stderr, e := cmd.StderrPipe()
	if !handleErr(e) {
		return
	}
	col := randomColor()
	colorize := color.New(col).FprintFunc()

	go func(stdout, stderr io.Reader, colorize func(w io.Writer, a ...interface{})) {
		go func() {
			reader := bufio.NewReader(stdout)
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}

				colorize(os.Stdout, line)
			}
		}()

		go func() {
			reader := bufio.NewReader(stderr)
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}

				colorize(os.Stderr, line)
			}
		}()
	}(stdout, stderr, colorize)


	colorize(os.Stdout, fmt.Sprintf("[%s] Starting: ", time.Now().Format("15:04:05")), command, "\n")
	err := cmd.Start()
	if !handleErr(e) {
		return
	}

	err = cmd.Wait()
	if !handleErr(err) {
		return
	}

	p.wg.Done()

	colorize(os.Stdout, fmt.Sprintf("[%s] Finished: ", time.Now().Format("15:04:05")), command, "\n")
}

func (p *play) printErrors() {
	for er := range p.errorChan {
		log.Print("Error during execution: ", er)
	}
}

func (p *play) Run() {
	go p.printErrors()

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

	if len(os.Args) < 2 {
		log.Print("Specify file to run")
		return
	}

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
