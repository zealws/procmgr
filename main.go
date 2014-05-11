package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	PrintStacktrace = true
)

type Pipe struct {
	out   io.ReadWriter
	color int
}

func (p Pipe) Read(arr []byte) (int, error) {
	return p.out.Read(arr)
}

func (p Pipe) Write(arr []byte) (int, error) {
	p.out.Write([]byte(fmt.Sprintf("\033[3%dm", p.color)))
	n, err := p.out.Write(arr)
	p.out.Write([]byte("\033[39m"))
	return n, err
}

func parseAfter(wait string) (string, string) {
	n, t := wait, "finished"
	if strings.Contains(wait, ".") {
		words := strings.Split(wait, ".")
		n = words[0]
		t = words[1]
	}
	return n, t
}

func waitForDependencies(ievents <-chan Event, config ProcessConfig) string {
	waitingFor := make(map[string]string)
	for _, wait := range config.After {
		n, t := parseAfter(wait)
		waitingFor[n] = t
	}
	for {
		select {
		case e := <-ievents:
			if t, ok := waitingFor[e.Id()]; ok {
				switch evt := e.(type) {
				case BeginEvent:
					if t == "started" {
						delete(waitingFor, evt.Id())
					}
				case EndEvent:
					if t == "finished" {
						delete(waitingFor, evt.Id())
					}
				case FailEvent:
					return evt.Id()
				}
			}
		default:
		}
		if len(waitingFor) == 0 {
			break
		}
	}
	return ""
}

func run(ievents <-chan Event, oevents chan<- Event, stdin, stdout, stderr Pipe, config ProcessConfig) {
	defer close(oevents)
	for {
		failed := waitForDependencies(ievents, config)
		if failed != "" {
			oevents <- FailEvent{config.Name, NewError("Dependency failed: "+failed, nil)}
			return
		}
		oevents <- BeginEvent(config.Name)
		cmd := exec.Command("sh", "-c", config.Command)
		for _, stream := range config.Streams {
			if stream == "stdin" {
				cmd.Stdin = stdin
			} else if stream == "stdout" {
				cmd.Stdout = stdout
			} else if stream == "stderr" {
				cmd.Stderr = stderr
			}
		}
		err := cmd.Run()
		if err != nil {
			oevents <- FailEvent{config.Name, NewError("Could not run command: "+config.Command, err)}
			return
		}
		if !config.Restart {
			break
		}
	}
	oevents <- EndEvent(config.Name)
	return
}

func handleProcesses(config *Config) error {
	oevents := make(map[string]chan Event)
	ievents := make(map[string]chan Event)
	broadcast := func(evt Event) {
		for _, ch := range ievents {
			ch <- evt
		}
	}
	finished := make(map[string]bool)

	failed := false

	events := make(chan Event, 10)
	for _, process := range config.Processes {
		oevents[process.Name] = make(chan Event, 10)
		ievents[process.Name] = make(chan Event, 10)
		finished[process.Name] = false
		go func(name string) {
			ch := oevents[name]
			for {
				c, ok := <-ch
				if !ok {
					break
				}
				events <- c
			}
		}(process.Name)
		go run(ievents[process.Name], oevents[process.Name], Pipe{os.Stdin, 0}, Pipe{os.Stdout, 2}, Pipe{os.Stderr, 1}, process)
	}

	for {
		select {
		case e := <-events:
			switch evt := e.(type) {
			default:
				log.Fatalf("Unknown event type %T: %+v\n", evt, evt)
			case BeginEvent:
				broadcast(e)
			case EndEvent:
				broadcast(e)
				finished[evt.Id()] = true
			case FailEvent:
				broadcast(e)
				finished[evt.Id()] = true
				failed = true
				fmt.Printf("%s failed: %s", evt.Id(), evt.Err.FullFormat())
			}
		default:
		}
		allDone := true
		for _, done := range finished {
			if !done {
				allDone = false
				break
			}
		}
		if allDone {
			if failed {
				return NewError("One or more processes failed", nil)
			}
			return nil
		}
	}
}

func CliMain(c *cli.Context) {
	PrintStacktrace = c.Bool("debug")
	config, err := ParseConfig(c.String("config"))
	handle(err)
	err = validateConfig(config)
	handle(err)
	err = handleProcesses(config)
	handle(err)
}

func handle(err error) {
	if err != nil {
		switch err := err.(type) {
		case Error:
			log.Fatal(err.FullFormat())
		default:
			log.Fatal(err)
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "procmgr"
	app.Usage = "Simple process manager."
	app.Flags = []cli.Flag{
		cli.StringFlag{"config, c", "procmgr.yaml", "path to config file"},
		cli.BoolFlag{"debug, d", "whether or not to print stacktraces"},
	}
	app.Action = CliMain
	app.Run(os.Args)
}
