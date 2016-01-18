// Par runs commands given as arguments in parallel. The exit code is the first
// non-zero exit code from any command or 1 in case of internal error.
package main

import (
	"bufio"
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

var sflag = flag.Bool("s", false, "read commands from stdin")

func main() {
	flag.Parse()

	log.SetFlags(0)
	log.SetPrefix("par: ")

	env := os.Environ()
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var cmds []*exec.Cmd
	if *sflag {
		s, err := os.Stdin.Stat()
		if err == nil && s.Mode()&os.ModeCharDevice != 0 {
			err = errors.New("stdin is empty")
		}
		if err != nil {
			log.Fatal(err)
		}
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			cmd := exec.Command("sh", "-c", sc.Text())
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = env
			cmd.Dir = dir
			cmds = append(cmds, cmd)
		}
		if sc.Err() != nil {
			log.Fatal(err)
		}
	} else {
		// TODO(pxi): duplicate stdin for each command.
		arg := os.Args[1:]
		cmds = make([]*exec.Cmd, len(arg))
		for i, arg := range arg {
			cmds[i] = exec.Command("sh", "-c", arg)
			cmds[i].Stdout = os.Stdout
			cmds[i].Stderr = os.Stderr
			cmds[i].Env = env
			cmds[i].Dir = dir
		}
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
