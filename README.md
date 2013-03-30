Introduction
============

gostress2 runs Go tests repeatedly to expose intermittent failures.

Examples
========

./bin/gostress2 -workers=8 -list=std -reps=100 -race=false -rebuild

./bin/gostress2 -list=std -duration=5m -race -rebuild

./bin/gostress2 -list=all -duration=5m -race -rebuild

GOPATH=$HOME/gopath ./bin/gostress2 -list=code.google.com/...
