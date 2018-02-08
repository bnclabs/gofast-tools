#! /usr/bin/env bash

go build;
./verify  -buffersize 512 -routines 100 -streams 10 -count 100000 -repeat 1
