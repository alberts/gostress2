package main

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

var globalEnv = []string{
	"TMPDIR=" + os.TempDir(),
	"PATH=" + os.Getenv("PATH"),
	"GOPATH=" + os.Getenv("GOPATH"),
}

func run(dir string, arg ...string) error {
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Dir = dir
	cmd.Env = globalEnv
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func stdout(dir string, arg ...string) (string, error) {
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Dir = dir
	cmd.Env = globalEnv
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(stdout)), nil
}

func pipe(wg *sync.WaitGroup, buf *[]byte, p io.ReadCloser) {
	defer wg.Done()
	b := make([]byte, 4096)
	for {
		n, err := p.Read(b)
		*buf = append(*buf, b[:n]...)
		if err != nil {
			return
		}
	}
}

func stdouterr(dir string, extraEnv []string, arg ...string) (string, string, error) {
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Dir = dir
	var env []string
	env = append(env, globalEnv...)
	env = append(env, extraEnv...)
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}

	var stdout []byte
	var stderr []byte

	var wg sync.WaitGroup
	wg.Add(2)
	p, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	go pipe(&wg, &stdout, p)

	p, err = cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}
	go pipe(&wg, &stderr, p)

	err = cmd.Run()
	wg.Wait()

	return strings.TrimSpace(string(stdout)), strings.TrimSpace(string(stderr)), err
}
