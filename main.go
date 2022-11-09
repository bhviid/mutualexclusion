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

var(
	ownPort = flag.Int("oPort", 8080, "own port")
	nextPort = flag.Int("nPort", 8081, "port of the next client in the ring")
	first = flag.Bool("first", false, "...")
)

func main() {
	//arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	//ownPort := int32(arg1) + 5000 // go run . 0..1..2..
	flag.Parse()

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
	//nextPort := int32((ownPort + 1)%3)
	
	//var conn *grpc.ClientConn
	fmt.Printf("Attempting to dial port: %d\n", *nextPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%v", *nextPort), grpc.WithInsecure(), grpc.WithBlock())
	
	fmt.Printf("Dialed port: %d\n", *nextPort)
	
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}
	//grpc.withblock to wait for the other peers/clients to start.
	
	// Defer means: When this function returns, call this method (meaing, one main is done, close connection)
	defer conn.Close()
	
	//  Create new Client from generated gRPC code from proto
	p.nextClient = me.NewMutualexclusionClient(conn)
	
	go func() {
		for {
			if p.hasToken {
				fmt.Println("has token! :D")
			}
			if !p.wantsAccess {
				p.passToNextClient()
			} else if p.wantsAccess && p.hasToken {
				//entered critical section
				fmt.Println("entered critical section ")
				time.Sleep(5 * time.Second)
				p.wantsAccess = false
				//p.passToNextClient()
				fmt.Println("left critical section ")
			}
		}
	}()

	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		p.wantsAccess = true
	}

}

func (p *peer) ReceiveToken(ctx context.Context, in *me.Token) (*me.Reply, error) {
	p.hasToken = true
	return &me.Reply{}, nil
}

func (p *peer) passToNextClient() {
	p.hasToken = false
	time.Sleep(2 * time.Second)
	p.nextClient.ReceiveToken(p.ctx, &me.Token{})

}
