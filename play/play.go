package play

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/olekukonko/tablewriter"
)

type Stage struct {
	Steps []string `yaml:"steps"`
}

type Play struct {
	Vars      map[string]string `yaml:"vars"`
	wg        *sync.WaitGroup
	Stages    []Stage `yaml:"play"`
	errorChan chan error

	tasks []*task
}

func NewPlay(stages []Stage, vars map[string]string) *Play {
	wg := new(sync.WaitGroup)
	errorChan := make(chan error)

	return &Play{
		wg:        wg,
		Vars:      vars,
		Stages:    stages,
		errorChan: errorChan,
	}
}

func (p *Play) printErrors() {
	for er := range p.errorChan {
		log.Print("Error during execution: ", er)
	}
}

func (p *Play) Run(verbose bool) {
	go p.printErrors()

	for stageIdx, stage := range p.Stages {
		p.wg.Add(len(stage.Steps))
		for taskIdx, command := range stage.Steps {
			t := newTask(fmt.Sprintf("%d_%d", stageIdx, taskIdx), command, p)
			p.tasks = append(p.tasks, t)

			go t.Run(verbose)
		}
		p.wg.Wait()
	}
}

func (p *Play) getLogsDir() string {
	return "/tmp/" + binaryName + "_log"
}

func (p *Play) DumpLogs() {
	for _, t := range p.tasks {
		err := t.DumpOutput(p.getLogsDir())
		if err != nil {
			log.Print(err)
		}
	}
}

func (p *Play) PrintResults() {
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
			strings.ReplaceAll(strings.Join(t.Cmd.Args, " "), "bash -c", ""),
			p.getLogsDir() + "/" + t.Name + "/full.log",
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Name", "Start", "End", "Duration", "Status", "Command", "Logs at"})
	_ = table.Append(data)
	_ = table.Render()
}
