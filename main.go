package main

import (
	"bufio"
	"context"
	me "dsysMe/proto"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"flag"

	"google.golang.org/grpc"
)

type peer struct {
	me.UnimplementedMutualexclusionServer
	id          int32
	nextClient  me.MutualexclusionClient
	ctx         context.Context
	hasToken    bool
	wantsAccess bool
}
//Flags for the command line when connecting a peer
var(
	ownPort = flag.Int("oPort", 8080, "own port")
	nextPort = flag.Int("nPort", 8081, "port of the next client in the ring")
	first = flag.Bool("first", false, "marks the process of the first in the logical ring")
)

func main() {
	flag.Parse()

	//https://stackoverflow.com/a/19966217
	f, err := os.OpenFile("OuputFileJustForYou.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &peer{
		id:  int32(*ownPort),
		ctx: ctx,
		hasToken: *first,
	}

	list, err := net.Listen("tcp", fmt.Sprintf(":%v", *ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %v: %v", ownPort, err)
	}

	grpcServer := grpc.NewServer()
	me.RegisterMutualexclusionServer(grpcServer, p)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()
	
	log.Printf("Node on port: %d -- Attempting to dial port: %d\n",*ownPort, *nextPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%v", *nextPort), grpc.WithInsecure(), grpc.WithBlock())
	
	log.Printf("Node on port: %d -- Dialed port: %d\n",*ownPort, *nextPort)
	
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}
	defer conn.Close()
	
	p.nextClient = me.NewMutualexclusionClient(conn)
	
	// Goroutine that will keep on passing the token around OR 
	// enter the critical zone if it has a token and a use for it.
	go func() {
		for {
			if p.hasToken && !p.wantsAccess {
				p.passToNextClient()
				log.Printf("Node on port: %d -- Passed token to %d\n",*ownPort, *nextPort)
			} else if p.wantsAccess && p.hasToken {
				//entered critical section
				log.Printf("Node on port: %d -- entered critical section\n", *ownPort)
				
				// let's simulate a hard working process by sleeping for 2 sec.
				time.Sleep(2 * time.Second)
				p.wantsAccess = false

				log.Printf("Node on port: %d -- left critical section\n",*ownPort)
			}
		}
	}()

	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		p.wantsAccess = true
	}

}

// What will I (the node) do when another node calls me?
func (p *peer) ReceiveToken(ctx context.Context, in *me.Token) (*me.Reply, error) {
	p.hasToken = true
	return &me.Reply{}, nil
}

// What will I (the node) do before/to call other peers.
func (p *peer) passToNextClient() {
	p.hasToken = false
	p.nextClient.ReceiveToken(p.ctx, &me.Token{})
}
