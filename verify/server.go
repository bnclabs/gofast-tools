package main

import "time"
import "sync"
import "log"
import "net"
import "fmt"
import "math/rand"

import gf "github.com/bnclabs/gofast"
import _ "github.com/bnclabs/gofast/http"

var mu sync.Mutex
var n_trans = make([]*gf.Transport, 0, 100)

func cleantrans() {
	tick := time.NewTicker(1 * time.Second)
	for {
		<-tick.C
		func() {
			mu.Lock()
			defer mu.Unlock()

			newtrans := make([]*gf.Transport, 0, 100)
			for _, trans := range n_trans {
				if trans.IsClosed() == false {
					newtrans = append(newtrans, trans)
				}
			}
			n_trans = newtrans
		}()
	}
}

func server(routines int) {
	go cleantrans()

	lis, err := net.Listen("tcp", options.addr)
	if err != nil {
		panic(fmt.Errorf("listen failed %v", err))
	}
	fmt.Printf("listening on %v\n", options.addr)

	ver := testVersion(1)

	opqstart, opqend := uint64(1000), uint64(1000+(routines*3))
	batchsize := uint64(rand.Intn(routines-1) + 1)
	setts := newsetts(uint64(options.buffersize), batchsize, opqstart, opqend)
	conncount := 0

	cmds := []int{sendpost, sendreq, sendstream}

	payload := rand.Intn(options.buffersize)
	msgstream := &msgStream{data: make([]byte, payload)}
	for i := 0; i < payload; i++ {
		msgstream.data[i] = 'a'
	}
	msgdone := newMsgDone([]byte("done"))

	var wg sync.WaitGroup

	for {
		if conn, err := lis.Accept(); err == nil {
			name := fmt.Sprintf("server-%v", conncount)
			conncount++
			trans, err := gf.NewTransport(name, conn, &ver, setts)
			if err != nil {
				panic("NewTransport server failed")
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
			trans.SubscribeMessage(
				&msgDone{},
				func(s *gf.Stream, msg gf.BinMessage) gf.StreamCallback {
					go func() {
						wg.Wait()
						if err := s.Response(msgdone, true); err != nil {
							log.Fatal(err)
						}
					}()
					return nil
				})
			trans.SubscribeMessage(
				&msgStat{},
				func(s *gf.Stream, msg gf.BinMessage) gf.StreamCallback {
					stat := make(map[string]uint64)
					for k, v := range trans.Stat() {
						stat[k] = v
					}
					respmsg := newMsgStat(stat)
					if err := s.Response(respmsg, true); err != nil {
						log.Fatal(err)
					}
					return nil
				})
			if err := trans.Handshake(); err != nil {
				panic(err)
			}

			func() {
				mu.Lock()
				defer mu.Unlock()
				n_trans = append(n_trans, trans)
			}()

			doers := make([]chan int, 0, 100)
			wg.Add(routines)
			for j := 0; j < routines; j++ {
				doch := make(chan int, options.count/routines)
				name := fmt.Sprintf("server-%v", j)
				go worker(name, nil, trans, doch, &wg)
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
		}
	}
}
