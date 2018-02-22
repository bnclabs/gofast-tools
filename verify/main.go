package main

import "flag"
import "runtime"
import "os"
import "math"
import "fmt"
import "log"
import "time"
import "runtime/pprof"
import "net/http"
import _ "net/http/pprof"

import "github.com/prataprc/color"
import golog "github.com/bnclabs/golog"
import "github.com/bnclabs/gofast"
import s "github.com/bnclabs/gosettings"

var _ = fmt.Sprintf("dummy")

var options struct {
	addr       string
	buffersize int
	conns      int
	routines   int
	streams    int
	count      int
	repeat     int
}

func argParse() {
	// generic options
	flag.StringVar(&options.addr, "addr", "127.0.0.1:9998",
		"server address")
	flag.IntVar(&options.buffersize, "buffersize", 1024*1024,
		"maximum buffer size")
	flag.IntVar(&options.conns, "conns", 16,
		"number of connections to launch")
	flag.IntVar(&options.routines, "routines", 10,
		"number of server/client routines to spawn")
	flag.IntVar(&options.streams, "streams", 100,
		"number of messages to tx per stream")
	flag.IntVar(&options.count, "count", 10000,
		"run count number of times.")
	flag.IntVar(&options.repeat, "repeat", 1,
		"repeat verification.")

	flag.Parse()
}

func main() {
	argParse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	golog.SetLogger(nil, s.Settings{"log.level": "warn", "log.file": ""})
	gofast.LogComponents("all")

	// start cpu profile, always enabled.
	fname := "verify.pprof"
	fd, err := os.Create(fname)
	if err != nil {
		log.Fatalf("unable to create %q: %v\n", fname, err)
	}
	defer fd.Close()
	pprof.StartCPUProfile(fd)
	defer pprof.StopCPUProfile()

	// start http server
	go func() {
		log.Println(http.ListenAndServe(":8080", nil))
	}()

	go server(options.routines)
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("number of connections %v\n", options.conns)
	for i := 0; i < options.repeat; i++ {
		fmt.Printf("repeat %v\n", i)
		cstat, sstat := client(options.routines)
		if rv := verify(cstat, sstat); rv > 0 {
			mkerror("Failed ...")
			os.Exit(1)
		}
	}

	// take memory profile.
	fname = "verify.mprof"
	fd, err = os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()
	pprof.WriteHeapProfile(fd)
}

func verify(cstat, sstat map[string]uint64) (rv int) {
	if x, y := cstat["n_rx"], sstat["n_tx"]; x != y {
		mkerror("mismatch client.n_rx server.n_tx %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_tx"], sstat["n_rx"]; x != y {
		mkerror("mismatch client.n_tx, server.n_rx %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_rxbyte"], sstat["n_txbyte"]; x != y {
		mkerror("mismatch client.n_rxbyte, server.n_txbyte %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_txbyte"], sstat["n_rxbyte"]; x != y {
		mkerror("mismatch client.n_txbyte, server.n_rxbyte %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_rxfin"], sstat["n_txfin"]; x != y {
		mkerror("mismatch client.n_rxfin, server.n_txfin %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_txfin"], sstat["n_rxfin"]; x != y {
		mkerror("mismatch client.n_txfin, server.n_rxfin %v,%v", x, y)
		rv++
	}
	x, y := cstat["n_rxpost"], sstat["n_txpost"]
	if diff := math.Abs(float64(x - y)); diff == 1 {
		fmt.Println("skipped a post, most likely heartbeat")
	} else if diff > 1 {
		mkerror("mismatch client.n_rxpost, server.n_txpost %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_txpost"], sstat["n_rxpost"]; x != y {
		mkerror("mismatch client.n_txpost, server.n_rxpost %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_rxreq"], sstat["n_txreq"]; x != y {
		mkerror("mismatch client.n_rxreq, server.n_txreq %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_txreq"], sstat["n_rxreq"]; x != y {
		mkerror("mismatch client.n_txreq, server.n_rxreq %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_rxresp"], sstat["n_txresp"]; x != y {
		mkerror("mismatch client.n_rxresp, server.n_txresp %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_txresp"], sstat["n_rxresp"]; x != y {
		mkerror("mismatch client.n_txresp, server.n_rxresp %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_rxstart"], sstat["n_txstart"]; x != y {
		mkerror("mismatch client.n_rxstart, server.n_txstart %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_txstart"], sstat["n_rxstart"]; x != y {
		mkerror("mismatch client.n_txstart, server.n_rxstart %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_rxstream"], sstat["n_txstream"]; x != y {
		mkerror("mismatch client.n_rxstream, server.n_txstream %v,%v", x, y)
		rv++
	}
	if x, y := cstat["n_txstream"], sstat["n_rxstream"]; x != y {
		mkerror("mismatch client.n_txstream, server.n_rxstream %v,%v", x, y)
		rv++
	}
	return
}

func mkerror(format string, args ...interface{}) {
	color.Red(fmt.Sprintf("Error "+format, args...))
}
