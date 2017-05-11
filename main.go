// main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/hashicorp/raft"
)

func main() {
	nodes := make(map[string]*raft.Raft, 5)
	peers := []string{}
	for i := 0; i < 5; i++ {
		addr, node, err := makeNode()
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
			fmt.Printf("%s: %v\n", addr, node.State())
		}
		time.Sleep(3 * time.Second)
	}
	log.Println("Finish")
}

func makeNode() (addr string, node *raft.Raft, err error) {
	// директория
	dir, err := ioutil.TempDir("", "raft")
	if err != nil {
		return addr, node, err
	}
	// конфиг
	conf := raft.DefaultConfig()
	// log store
	fsm := &MockFSM{}
	// transport
	addr, trans := raft.NewInmemTransport("")
	//
	store := raft.NewInmemStore()
	//
	snap, err := raft.NewFileSnapshotStore(dir, 3, nil)
	if err != nil {
		return addr, node, err
	}
	//
	peers := raft.NewJSONPeers(dir, trans)
	fmt.Printf("Created %s: %d\n", addr, dir)

	node, err = raft.NewRaft(conf, fsm, store, store, snap, peers, trans)
	if err != nil {
		return addr, node, err
	}

	return addr, node, nil
}
