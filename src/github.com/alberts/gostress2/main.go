package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type work struct {
	Package
	revision string
	reps     int
	timeLeft time.Duration
	done     chan struct{}
}

func (w *work) testCmd() string {
	return filepath.Join(w.Dir, path.Base(w.ImportPath)+".test")
}

func (w *work) Setup() error {
	if w.revision == "" {
		revision, err := stdout(w.Dir, "hg", "log", "--follow", "--limit=1", "--template={node}")
		if err != nil {
			return err
		}
		w.revision = revision
		//log.Printf("%s: revision %s", w.ImportPath, w.revision)
	}
	testCmd := w.testCmd()
	//log.Printf("%s: %s", w.ImportPath, testCmd)
	if _, err := os.Stat(testCmd); err == nil && !*rebuild {
		return nil
	}
	gotest := []string{"go", "test", "-c"}
	if *race {
		gotest = append(gotest, "-race")
	}
	if err := run(w.Dir, gotest...); err != nil {
		return err
	}
	return nil
}

func (w *work) GOMAXPROCS() string {
	switch rand.Intn(4) {
	case 0:
		return ""
	case 1:
		return "GOMAXPROCS=1"
	case 2:
		return "GOMAXPROCS=2"
	case 3:
		return "GOMAXPROCS=" + strconv.Itoa(1+rand.Intn(1024))
	}
	panic("GOMAXPROCS: invalid case")
}

func (w *work) GOGC() (string, bool) {
	switch rand.Intn(5) {
	case 0:
		return "", true
	case 1:
		return "GOGC=off", false
	case 2:
		return "GOGC=1", true
	case 3:
		return "GOGC=100", true
	case 4:
		return "GOGC=" + strconv.Itoa(1+rand.Intn(100)), true
	}
	panic("GOGC: invalid case")
}

func (w *work) GOGCTRACE() string {
	switch rand.Intn(3) {
	case 0:
		return ""
	case 1:
		return "GOGCTRACE=0"
	case 2:
		return "GOGCTRACE=1"
	}
	panic("GOGCTRACE: invalid case")
}

func (w *work) GOTRACEBACK() string {
	switch rand.Intn(5) {
	case 0:
		return ""
	case 1:
		return "GOTRACEBACK=0"
	case 2:
		return "GOTRACEBACK=1"
	case 3:
		return "GOTRACEBACK=2"
	case 4:
		return "GOTRACEBACK=crash"
	}
	panic("GOTRACEBACK: invalid case")
}

func (w *work) sudo() []string {
	switch rand.Intn(2) {
	case 0:
		return nil
	case 1:
		return []string{"sudo", "-E"}
	}
	panic("sudo: invalid case")
}

func (w *work) strace() []string {
	switch rand.Intn(2) {
	case 0:
		return nil
	case 1:
		return []string{"strace", "-f", "-q", "-o/dev/null"}
	}
	panic("strace: invalid case")
}

var slowTests = map[string]struct{}{
	"archive/zip": struct{}{},
	"math/big":    struct{}{},
	"net":         struct{}{},
	"net/http":    struct{}{},
	"regexp":      struct{}{},
}

func (w *work) testCpu() string {
	// don't run slow tests more than once
	if _, ok := slowTests[w.ImportPath]; ok {
		return ""
	}
	maxLength := 9
	if *race {
		maxLength = 3
	}
	length := rand.Intn(maxLength)
	if length == 0 {
		return ""
	}
	var args []string
	for i := 0; i < length; i++ {
		cpu := 1 + rand.Intn(256)
		args = append(args, strconv.Itoa(cpu))
	}
	return "-test.cpu=" + strings.Join(args, ",")
}

func (w *work) testShort() string {
	if rand.Intn(2) == 0 {
		return ""
	}
	return "-test.short"
}

func (w *work) testVerbose() string {
	if rand.Intn(2) == 0 {
		return ""
	}
	return "-test.v"
}

func (w *work) testBench() string {
	if rand.Intn(2) == 0 {
		return ""
	}
	return "-test.bench=."
}

func appendStr(v []string, s ...string) []string {
	if len(s) == 0 {
		return v
	}
	if len(s) == 1 && s[0] == "" {
		return v
	}
	return append(v, s...)
}

func (w *work) Do() {
	t0 := time.Now()
	defer close(w.done)
	defer func() {
		w.reps--
		w.timeLeft -= time.Now().Sub(t0)
		if w.timeLeft < 0 {
			w.timeLeft = 0
		}
		log.Printf("%s: remaining: %d repetitions, %v time", w.ImportPath, w.reps, w.timeLeft)
	}()

	var env []string
	env = appendStr(env, w.GOMAXPROCS())
	gogc, hasgc := w.GOGC()
	env = appendStr(env, gogc)
	env = appendStr(env, w.GOGCTRACE())
	env = appendStr(env, w.GOTRACEBACK())

	var args []string
	if *sudo {
		args = appendStr(args, w.sudo()...)
	}
	if *strace {
		args = appendStr(args, w.strace()...)
	}
	args = appendStr(args, w.testCmd())
	if hasgc {
		args = appendStr(args, w.testCpu())
	}
	if hasgc {
		args = appendStr(args, w.testShort())
	} else {
		args = append(args, "-test.short")
	}
	args = appendStr(args, w.testVerbose())
	if hasgc {
		args = appendStr(args, w.testBench())
	}

	log.Printf("%s: %s %s", w.ImportPath, env, args)

	so, se, err := stdouterr(w.Dir, env, args...)
	if err != nil {
		log.Printf("%s: %v", w.ImportPath, err)
		println(so)
		println(se)
	}
}

func (w *work) Done() bool {
	return w.reps <= 0 || w.timeLeft <= 0
}

var seed = flag.Int64("seed", time.Now().UnixNano(), "seed")
var list = flag.String("list", "std", "packages to test")
var workers = flag.Int("workers", 1, "number of workers")
var race = flag.Bool("race", false, "use race detector")
var rebuild = flag.Bool("rebuild", false, "rebuild tests")
var reps = flag.Int("reps", 1, "repetitions")
var duration = flag.Duration("duration", 1*time.Minute, "duration")
var strace = flag.Bool("strace", false, "strace some tests")
var sudo = flag.Bool("sudo", false, "sudo some tests")

func do(q <-chan *work, done <-chan struct{}) {
	for {
		select {
		case w, ok := <-q:
			if !ok {
				return
			}
			if err := w.Setup(); err != nil {
				log.Printf("%s: %v\n", w.ImportPath, err)
				continue
			}
			w.Do()
		case <-done:
			return
		}
	}
}

func main() {
	flag.Parse()

	log.Printf("seed=%d", *seed)
	rand.Seed(*seed)

	// get randomly shuffled list of packages
	packages := getPackages(*list, true)

	// start workers
	q := make(chan *work)
	done := make(chan struct{})
	var workersWg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()
			do(q, done)
		}()
	}

	var wg sync.WaitGroup
	for _, pkg := range packages {
		wg.Add(1)
		w := &work{
			Package:  pkg,
			reps:     *reps,
			timeLeft: *duration,
		}
		go func() {
			defer wg.Done()
			for !w.Done() {
				w.done = make(chan struct{})
				select {
				case q <- w:
				case <-done:
					return
				}
				<-w.done
			}
		}()
	}
	wg.Wait()
	close(q)
	workersWg.Wait()
}
