package main

import "sync"
import "log"
import "time"
import "net"
import "fmt"
import "math/rand"

import gf "github.com/prataprc/gofast"

func client(routines int) (cstat, sstat map[string]uint64) {
	var wg sync.WaitGroup
	var av = &averageInt64{}

	now := time.Now()

	ver := testVersion(1)
	n_trans := make([]*gf.Transport, 0, options.conns)
	cmds := []int{sendpost, sendreq, sendstream}

	payload := rand.Intn(options.buffersize)
	msgstream := &msgStream{data: make([]byte, payload)}
	for i := 0; i < payload; i++ {
		msgstream.data[i] = 'a'
	}
	msgdone := newMsgDone([]byte("done"))
	msgstat := newMsgStat(map[string]uint64{})

	for i := 0; i < options.conns; i++ {
		wg.Add(routines)
		go func(n int) {
			opqstart, opqend := uint64(16000), uint64(16000+(routines*3))
			batchsize := uint64(rand.Intn(routines-1) + 1)
			name := fmt.Sprintf("verify-client-%v", n)
			setts := newsetts(uint64(options.buffersize), batchsize, opqstart, opqend)
			setts["tags"] = ""
			conn, err := net.Dial("tcp", options.addr)
			if err != nil {
				panic(err)
			}
			trans, err := gf.NewTransport(name, conn, &ver, setts)
			if err != nil {
				panic(err)
			}

			trans.FlushPeriod(50 * time.Millisecond)
			//trans.SendHeartbeat(5 * time.Second)
			trans.SubscribeMessage(
				&msgPost{},
				func(s *gf.Stream, msg gf.BinMessage) gf.StreamCallback {
					return nil
				})
			trans.SubscribeMessage(
				&msgReqsp{},
				func(s *gf.Stream, msg gf.BinMessage) gf.StreamCallback {
					var rxmsg msgReqsp
					rxmsg.Decode(msg.Data)
					if err := s.Response(&rxmsg, true); err != nil {
						log.Fatal(err)
					}
					return nil
				})
			trans.SubscribeMessage(
				&msgStreamStart{},
				func(s *gf.Stream, msg gf.BinMessage) gf.StreamCallback {
					var wg sync.WaitGroup
					wg.Add(1)
					go func() {
						for j := 0; j < rand.Intn(options.streams); j++ {
							err := s.Stream(msgstream, true)
							if err != nil {
								log.Printf("error stream: %v\n", err)
								break
							}
						}
						wg.Wait()
						s.Close()
					}()
					return func(rxstrmsg gf.BinMessage, ok bool) {
						if ok && rxstrmsg.ID != msgstream.ID() {
							log.Printf("unexp. message %T %v\n", rxstrmsg.ID, rxstrmsg)
						} else if ok == false {
							wg.Done()
						}
					}
				})
			trans.SubscribeMessage(
				&msgStream{},
				func(s *gf.Stream, msg gf.BinMessage) gf.StreamCallback {
					return nil
				})
			trans.SubscribeMessage(&msgDone{}, nil)
			trans.SubscribeMessage(&msgStat{}, nil)
			if err := trans.Handshake(); err != nil {
				panic(err)
			}

			n_trans = append(n_trans, trans)

			doers := make([]chan int, 0, 100)
			for j := 0; j < routines; j++ {
				doch := make(chan int, options.count/routines)
				name := fmt.Sprintf("client-%v", j)
				go worker(name, av, trans, doch, &wg)
				doers = append(doers, doch)
			}

			go func(trans *gf.Transport, doers []chan int) {
				numdoers := len(doers)
				for k := 0; k < options.count; k++ {
					doers[k%numdoers] <- cmds[rand.Intn(len(cmds))]
				}
				for _, doch := range doers {
					close(doch)
				}
			}(trans, doers)
		}(i)
	}
	wg.Wait()

	// handshake for end
	for _, trans := range n_trans {
		trans.Request(msgdone, true, nil)
	}

	// get server stats
	sstat = make(map[string]uint64)
	for _, trans := range n_trans {
		var rmsg msgStat

		y0 := trans.Stat()["n_rxbyte"]
		if err := trans.Request(msgstat, true, &rmsg); err != nil {
			fmt.Printf("%v\n", err)
			panic("exit")
		}
		y1 := trans.Stat()["n_rxbyte"]

		stat := rmsg.data
		stat["n_tx"] += 1
		stat["n_rx"] += 1
		stat["n_txbyte"] += y1 - y0
		stat["n_rxreq"] += 1
		stat["n_txresp"] += 1
		for k, v := range stat {
			if val, ok := sstat[k]; ok {
				sstat[k] = val + v
			} else {
				sstat[k] = v
			}
		}
	}

	// get client stats
	cstat = make(map[string]uint64)
	for _, trans := range n_trans {
		for k, v := range trans.Stat() {
			if val, ok := cstat[k]; ok {
				cstat[k] = val + v
			} else {
				cstat[k] = v
			}
		}
	}

	for _, trans := range n_trans {
		trans.Close()
	}

	n, m := av.samples(), time.Duration(av.mean())
	v, s := time.Duration(av.variance()), time.Duration(av.sd())
	fmsg := "client latency: n:%v mean:%v var:%v sd:%v\n"
	fmt.Printf(fmsg, n, m, v, s)
	fmt.Printf("took %v\n", time.Since(now))

	return cstat, sstat
}

const (
	sendpost int = iota + 1
	sendreq
	sendstream
)

func worker(
	name string, av *averageInt64, t *gf.Transport,
	doch chan int, wg *sync.WaitGroup) {

	payload := rand.Intn(options.buffersize)
	msgpost := &msgPost{data: make([]byte, payload)}
	for i := 0; i < payload; i++ {
		msgpost.data[i] = 'a'
	}
	msgreq := &msgReqsp{data: make([]byte, payload)}
	for i := 0; i < payload; i++ {
		msgreq.data[i] = 'a'
	}
	msgstart := &msgStreamStart{data: make([]byte, payload)}
	for i := 0; i < payload; i++ {
		msgstart.data[i] = 'a'
	}
	msgstream := &msgStream{data: make([]byte, payload)}
	for i := 0; i < payload; i++ {
		msgstream.data[i] = 'a'
	}

	for doit := range doch {
		since := time.Now()
		switch doit {
		case sendpost:
			if err := t.Post(msgpost, false); err != nil {
				fmt.Printf("%v\n", err)
				panic("exit")
			}

		case sendreq:
			var rmsg msgReqsp

			if err := t.Request(msgreq, false, &rmsg); err != nil {
				fmt.Printf("%v\n", err)
				panic("exit")
			}

		case sendstream:
			donech := make(chan struct{})
			stream, err := t.Stream(
				msgstart, false, func(rmsg gf.BinMessage, ok bool) {
					if ok == false {
						close(donech)
					} else if rmsg.ID != msgstream.ID() {
						log.Printf("unexpected message %v\n", rmsg.ID)
					}
				})
			if err != nil {
				log.Fatal(err)
			}
			for j := 0; j < rand.Intn(options.streams); j++ {
				if err := stream.Stream(msgstream, false); err != nil {
					log.Printf("error stream: %v\n", err)
					break
				}
			}
			stream.Close()
			<-donech
		}
		if av != nil {
			av.add(int64(time.Since(since)))
		}
	}
	if _, err := t.Whoami(); err != nil {
		log.Fatal(err)
	}
	wg.Done()
}
