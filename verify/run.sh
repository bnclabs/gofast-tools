#! /usr/bin/env bash

go build;
./verify  -buffersize 512 -routines 100 -streams 10 -count 100000 -repeat 100
go tool pprof --alloc_space -svg verify verify.mprof > alloc_space.svg
go tool pprof --inuse_space -svg verify verify.mprof > inuse_space.svg
