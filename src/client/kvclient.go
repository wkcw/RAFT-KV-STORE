package client

import (
	"google.golang.org/grpc"
	"flag"
	"fmt"
	pb "proto"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:9527", "The server address in the format of host:port")
)

func main() {
	conn, err := grpc.Dial(*serverAddr)
	if err != nil {
		fmt.Errorf("Connection Failed")
	}
	defer conn.Close()
	client := pb.key


}