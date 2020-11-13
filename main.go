package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/martinlindhe/notify"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

const TimeFormat = "15:04:05"

func randomColor() color.Attribute {
	return AvailableColors[rand.Intn(len(AvailableColors))]
}

func colorizeAndWrite(colorize func(w io.Writer, a ...interface{}), input io.Reader, out io.Writer, cb func(string)) {
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		colorize(out, line)
		cb(line)
	}
}

type stage struct {
	Steps []string `yaml:"steps"`
}

type play struct {
	wg        *sync.WaitGroup
	Stages    []stage `yaml:"play"`
	errorChan chan error

	tasks []*task
}

type task struct {
	Success    bool
	Stdout     string
	Stderr     string
	FullOutput string
	StartedAt  time.Time
	EndedAt    time.Time
	Cmd        *exec.Cmd
	p          *play
	Name       string
}

func (t *task) DumpOutput(to string) error {
	err := os.MkdirAll(to+"/"+t.Name, 0755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(to+"/"+t.Name+"/stdout.log", []byte(t.Stdout), 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(to+"/"+t.Name+"/stderr.log", []byte(t.Stderr), 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(to+"/"+t.Name+"/full.log", []byte(t.FullOutput), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (t *task) Run() {
	handleErr := func(err error) bool {
		if err != nil {
			t.p.errorChan <- err
			t.p.wg.Done()
			t.Success = false
			t.EndedAt = time.Now()

			notify.Notify(os.Args[0], "Command failed", fmt.Sprintf("%s exited with %+v", t.Cmd.Args, err), "")
			return false
		}

		return true
	}

	stdout, e := t.Cmd.StdoutPipe()
	if !handleErr(e) {
		return
	}

	stderr, e := t.Cmd.StderrPipe()
	if !handleErr(e) {
		return
	}
	col := randomColor()
	colorize := color.New(col).FprintFunc()

	saveStdout := func(out string) {
		t.Stdout = t.Stdout + out
		t.FullOutput = t.FullOutput + out
	}

	saveStderr := func(out string) {
		t.Stderr = t.Stderr + out
		t.FullOutput = t.FullOutput + out
	}

	go colorizeAndWrite(colorize, stdout, os.Stdout, saveStdout)
	go colorizeAndWrite(colorize, stderr, os.Stderr, saveStderr)

	t.StartedAt = time.Now()
	colorize(os.Stdout, fmt.Sprintf("[%s] Starting: ", t.StartedAt.Format(TimeFormat)), t.Cmd.Args, "\n")
	err := t.Cmd.Start()
	if !handleErr(e) {
		return
	}

	err = t.Cmd.Wait()
	if !handleErr(err) {
		return
	}

	t.p.wg.Done()

	t.EndedAt = time.Now()
	colorize(os.Stdout, fmt.Sprintf("[%s] Finished: ", t.EndedAt.Format(TimeFormat)), t.Cmd.Args, "\n")
	t.Success = true
}

func NewTask(name, command string, parent *play) *task {
	cmd := exec.Command("bash", "-c", command)
	return &task{
		StartedAt: time.Now(),
		EndedAt:   time.Now(),
		Cmd:       cmd,
		p:         parent,
		Name:      name,
	}
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

func (p *play) printErrors() {
	for er := range p.errorChan {
		log.Print("Error during execution: ", er)
	}
}

func (p *play) Run() {
	go p.printErrors()

	for stageIdx, stage := range p.Stages {
		p.wg.Add(len(stage.Steps))
		for taskIdx, command := range stage.Steps {
			t := NewTask(fmt.Sprintf("%d_%d", stageIdx, taskIdx), command, p)
			p.tasks = append(p.tasks, t)

			go t.Run()
		}
		p.wg.Wait()
	}
}

func (p *play) getLogsDir() string {
	return "/tmp/" + filepath.Base(os.Args[0]) + "_log"
}

func (p *play) DumpLogs() {
	for _, t := range p.tasks {
		err := t.DumpOutput(p.getLogsDir())
		if err != nil {
			log.Print(err)
		}
	}
}

func (p *play) PrintResults() {
	var data [][]string

	for _, t := range p.tasks {
		var status string
		if t.Success {
			status = "success"
		} else {
			status = "failed"
		}
		data = append(data, []string{
			t.Name,
			t.StartedAt.Format(TimeFormat),
			t.EndedAt.Format(TimeFormat),
			t.EndedAt.Sub(t.StartedAt).String(),
			status,
			strings.Replace(strings.Join(t.Cmd.Args, " "), "bash -c", "", -1),
			p.getLogsDir() + "/" + t.Name + "/full.log",
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Start", "End", "Duration", "Status", "Command", "Logs at"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
}

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		log.Print("Specify config path")
		log.Print("Usage: ", os.Args[0], " [config path]")
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
	log.Print("Finished run")

	pl.DumpLogs()
	pl.PrintResults()

}
