// main.go
package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

func main() {
	nodes := make(map[string]*raft.Raft, 5)
	peers := []string{}
	// конфиг
	conf := raft.DefaultConfig()
	conf.Logger = log.New(os.Stderr, "raft", log.Flags())
	// создаем ноды
	for i := 0; i < 5; i++ {
		addr, node, err := makeNode(conf)
		if err != nil {
			panic(err)
		}
		defer node.Shutdown()
		node.SetPeers(peers)
		peers = append(peers, addr)
		nodes[addr] = node
	}

	//
	for {
		for addr, node := range nodes {
			state := node.State()
			log.Printf("%s: %v\n", addr, state)
			if state == raft.Follower || state == raft.Leader {
				af := node.Apply([]byte("from "+state.String()), time.Second)
				log.Printf("%s APPLY: %+v\n", addr, af)
			}
			time.Sleep(3 * time.Second)
		}
	}
	log.Println("Finish")
}

func makeNode(conf *raft.Config) (addr string, node *raft.Raft, err error) {
	// директория
	dir, err := ioutil.TempDir("", "raft")
	if err != nil {
		return addr, node, err
	}
	// log store
	fsm := &MockFSM{}
	// transport
	trans, err := raft.NewTCPTransport("127.0.0.1:0", nil, 2, time.Second, nil)
	if err != nil {
		return addr, node, err
	}
	addr = trans.LocalAddr()
	//
	store := raft.NewInmemStore()
	//
	snap, err := raft.NewFileSnapshotStore(dir, 3, nil)
	if err != nil {
		return addr, node, err
	}
	//
	peers := raft.NewJSONPeers(dir, trans)
	log.Printf("Created %s: %d\n", addr, dir)

	node, err = raft.NewRaft(conf, fsm, store, store, snap, peers, trans)
	if err != nil {
		return addr, node, err
	}

	return addr, node, nil
}
