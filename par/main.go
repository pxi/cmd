// Par runs commands given as arguments in parallel. The exit code is the first
// non-zero exit code from any command or 1 in case of internal error.
package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

func main() {
	flag.Parse()

	log.SetFlags(0)
	log.SetPrefix("par: ")

	env := os.Environ()
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// TODO(pxi): duplicate stdin for each command.
	// TODO(pxi): read commands from stdin.
	args := os.Args[1:]
	cmds := make([]*exec.Cmd, len(args))
	for i, arg := range args {
		cmds[i] = exec.Command("sh", "-c", arg)
		cmds[i].Stdout = os.Stdout
		cmds[i].Stderr = os.Stderr
		cmds[i].Env = env
		cmds[i].Dir = dir
	}

	i, err := proc(run(cmds))
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(i)
}

func proc(ch <-chan error) (c int, err error) {
	for err := range ch {
		nc, nerr := unwind(err)
		if c == 0 {
			c = nc
		}
		if err == nil {
			err = nerr
		}
	}
	return
}

func unwind(err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
		// Most likely an platform where WaitStatus is not defined.
		err = errors.New("cannot read exit code")
	}

	return 0, err
}

func run(cmds []*exec.Cmd) <-chan error {
	wg := sync.WaitGroup{}
	ch := make(chan error)

	go func() {
		wg.Wait()
		close(ch)
	}()

	wg.Add(len(cmds))
	for _, cmd := range cmds {
		go func(cmd *exec.Cmd) {
			ch <- cmd.Run()
			wg.Done()
		}(cmd)
	}

	return ch
}
