package main

import (
	"bufio"
	"context"
	me "dsysMe/proto"
	"fmt"
	"log"
	"net"
	"os"
	
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
	f, err := os.OpenFile("bigoutput2.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
	
	go func() {
		for {
			if p.hasToken && !p.wantsAccess {
				p.passToNextClient()
				//log.Printf("Node on port: %d -- Passed token to %d\n",*ownPort, *nextPort)
			} else if p.wantsAccess && p.hasToken {
				//entered critical section
				log.Printf("Node on port: %d -- entered critical section\n", *ownPort)
				//time.Sleep(5 * time.Second)
				p.wantsAccess = false
				//p.passToNextClient()
				log.Printf("Node on port: %d -- left critical section\n",*ownPort)
			}
		}
	}()

	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		p.wantsAccess = true
	}

}

func (p *peer) ReceiveToken(ctx context.Context, in *me.Token) (*me.Reply, error) {
	//time.Sleep(2 * time.Second)
	p.hasToken = true
	return &me.Reply{}, nil
}

func (p *peer) passToNextClient() {
	p.hasToken = false
	p.nextClient.ReceiveToken(p.ctx, &me.Token{})
}
