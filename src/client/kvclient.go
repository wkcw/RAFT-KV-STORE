package client

import (
	"google.golang.org/grpc"
	"flag"
	"log"
	pb "proto"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:9527", "The server address in the format of host:port")
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(*serverAddr)
	if err != nil {
		log.Errorf("Connection Failed: %v", err)
	}
	defer conn.Close()
	// set up a new client
	c := pb.NewKeyValueStoreClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// put operation, need to be fixed
	r, err := c.Put(ctx, &pb.PutRequest{Key: "testKey", Value: "testValue"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Return code: %s", r.Ret)

	// get operation, need to be fixed
	r1, err1 := c.Get(ctx, &pb.GetRequest{Key: "testKey"})
	if err1 != nil {
		log.Fatalf("could not get: %v", err1)
	}
	log.Printf("Value: %s", r1.Value)
}