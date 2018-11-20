

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync/atomic"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/node"
	"github.com/ground-x/go-gxplatform/networks/p2p"
	"github.com/ground-x/go-gxplatform/networks/p2p/discover"
	"github.com/ground-x/go-gxplatform/networks/p2p/simulations"
	"github.com/ground-x/go-gxplatform/networks/p2p/simulations/adapters"
	"github.com/ground-x/go-gxplatform/networks/rpc"
	"time"
)

var adapterType = flag.String("adapter", "cnsim", `node adapter to use (one of "sim", "exec" or "docker")`)
var logger = log.NewModuleLogger("networks/p2p/simulations/cnism")

// main() starts a simulation network which contains nodes running a simple
// ping-pong protocol
func main() {
	flag.Parse()

	// set the log level to Trace
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	// register a single ping-pong service
	services := map[string]adapters.ServiceFunc{
		"cn-sim": func(ctx *adapters.ServiceContext) (node.Service, error) {
			return newCnSimService(ctx.Config.ID), nil
		},
	}
	adapters.RegisterServices(services)

	// create the NodeAdapter
	var adapter adapters.NodeAdapter

	switch *adapterType {

	case "cnsim":
		logger.Info("using cnsim adapter")
		adapter = adapters.NewCnAdapter(services)

	case "exec":
		tmpdir, err := ioutil.TempDir("", "p2p-example")
		if err != nil {
			logger.Crit("error creating temp dir", "err", err)
		}
		defer os.RemoveAll(tmpdir)
		logger.Info("using exec adapter", "tmpdir", tmpdir)
		adapter = adapters.NewExecAdapter(tmpdir)

	case "docker":
		logger.Info("using docker adapter")
		var err error
		adapter, err = adapters.NewDockerAdapter()
		if err != nil {
			logger.Crit("error creating docker adapter", "err", err)
		}

	default:
		logger.Crit(fmt.Sprintf("unknown node adapter %q", *adapterType))
	}

	// start the HTTP API
	logger.Info("starting simulation server on 0.0.0.0:8888...")
	network := simulations.NewNetwork(adapter, &simulations.NetworkConfig{
		DefaultService: "cn-sim",
	})
	if err := http.ListenAndServe(":8888", simulations.NewServer(network)); err != nil {
		logger.Crit("error starting simulation server", "err", err)
	}
}

// pingPongService runs a ping-pong protocol between nodes where each node
// sends a ping to all its connected peers every 10s and receives a pong in
// return
type cnSimService struct {
	id       discover.NodeID
	logger   log.Logger
	received int64
}

func newCnSimService(id discover.NodeID) *cnSimService {
	return &cnSimService{
		id:     id,
		logger: logger.NewWith("node.id", id),
	}
}

func (p *cnSimService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{{
		Name:     "cn-sim",
		Version:  1,
		Length:   2,
		Run:      p.Run,
		NodeInfo: p.Info,
	}}
}

func (p *cnSimService) APIs() []rpc.API {
	return nil
}

func (p *cnSimService) Start(server *p2p.Server) error {
	p.logger.Info("cn-sim service starting")
	return nil
}

func (p *cnSimService) Stop() error {
	p.logger.Info("cn-sim service stopping")
	return nil
}

func (p *cnSimService) Info() interface{} {
	return struct {
		Received int64 `json:"received"`
	}{
		atomic.LoadInt64(&p.received),
	}
}

const (
	pingMsgCode = iota
	pongMsgCode
)

// Run implements the ping-pong protocol which sends ping messages to the peer
// at 10s intervals, and responds to pings with pong messages.
func (p *cnSimService) Run(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
	// log := p.logger.New("peer.id", peer.ID())


	errC := make(chan error)

	go func() {
		for range time.Tick(10 * time.Second) {
			logger.Info("sending ping")
			if err := p2p.Send(rw, pingMsgCode, "PING"); err != nil {
				errC <- err
				return
			}
		}
	}()
	go func() {
		for {
			msg, err := rw.ReadMsg()
			if err != nil {
				errC <- err
				return
			}
			payload, err := ioutil.ReadAll(msg.Payload)
			if err != nil {
				errC <- err
				return
			}
			logger.Info("received message", "msg.code", msg.Code, "msg.payload", string(payload))
			atomic.AddInt64(&p.received, 1)
			if msg.Code == pingMsgCode {
				logger.Info("sending pong")
				go p2p.Send(rw, pongMsgCode, "PONG")
			}
		}
	}()

	return <-errC
}
