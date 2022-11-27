package play

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/martinlindhe/notify"
)

const TimeFormat = "15:04:05"

var binaryName = filepath.Base(os.Args[0])

func randomColor() color.Attribute {
	return AvailableColors[rand.Intn(len(AvailableColors))]
}

func colorizeAndWrite(prefix string, colorize func(w io.Writer, a ...interface{}), input io.Reader, out io.Writer, cb func(string)) {
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if len(prefix) > 0 {
			line = prefix + ": " + line
		}
		colorize(out, line)
		cb(line)
	}
}

type task struct {
	Success    bool
	Stdout     string
	Stderr     string
	FullOutput string
	StartedAt  time.Time
	EndedAt    time.Time
	Cmd        *exec.Cmd
	p          *Play
	Name       string
}

func (t *task) DumpOutput(to string) error {
	err := os.MkdirAll(to+"/"+t.Name, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(to+"/"+t.Name+"/stdout.log", []byte(t.Stdout), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(to+"/"+t.Name+"/stderr.log", []byte(t.Stderr), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(to+"/"+t.Name+"/full.log", []byte(t.FullOutput), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (t *task) Run(verbose bool) {
	handleErr := func(err error) bool {
		if err != nil {
			t.p.errorChan <- err
			t.p.wg.Done()
			t.Success = false
			t.EndedAt = time.Now()

			notify.Notify(binaryName, "Command failed", fmt.Sprintf("%s exited with %+v", t.Cmd.Args, err), "")
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

	prefix := ""
	if verbose {
		prefix = strings.Join(t.Cmd.Args, " ")
	}
	go colorizeAndWrite(prefix, colorize, stdout, os.Stdout, saveStdout)
	go colorizeAndWrite(prefix, colorize, stderr, os.Stderr, saveStderr)

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

	// Notify when long-running tasks finishes
	diffTime := t.EndedAt.Sub(t.StartedAt)
	if diffTime.Minutes() > 5 {
		msg := fmt.Sprintf("%s finished after %s", t.Cmd.Args, diffTime.String())
		notify.Notify(binaryName, "Command finished", msg, "")
	}
}

func newTask(name, command string, parent *Play) *task {
	t := template.Must(template.New("task").Parse(command))

	var commandBuilder strings.Builder
	e := t.Execute(&commandBuilder, parent.Vars)
	if e != nil {
		parent.errorChan <- e
		return nil
	}

	cmd := exec.Command("bash", "-c", commandBuilder.String())
	return &task{
		StartedAt: time.Now(),
		EndedAt:   time.Now(),
		Cmd:       cmd,
		p:         parent,
		Name:      name,
	}
}
