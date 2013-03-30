Introduction
============

gostress2 runs Go tests repeatedly to expose intermittent failures.

Usage
=====

```
Usage of gostress2:
  -duration=1m0s: duration
  -list="std": packages to test
  -race=false: use race detector
  -rebuild=false: rebuild tests
  -reps=1: repetitions
  -seed=1364657604277567273: seed
  -strace=false: strace some tests
  -sudo=false: sudo some tests
  -workers=1: number of workers
```

Examples
========

./bin/gostress2 -workers=8 -list=std -reps=100 -race=false -rebuild

./bin/gostress2 -list=std -duration=5m -race -rebuild

./bin/gostress2 -list=all -duration=5m -race -rebuild

GOPATH=$HOME/gopath ./bin/gostress2 -list=code.google.com/...
