package main

import (
	pb "chaosmonkey"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"time"
	"util"
)


func main() {
	config := util.CreateConfig()
	serverlist := config.ServerList
	for {

		fmt.Printf("%v\n", time.Now())
		for _, server := range serverlist.Servers{
			var address string = server.Addr
			// connect to the server
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()
			c := pb.NewChaosMonkeyClient(conn)

			// Contact the server and print out its response.
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			r1, err1 := c.UpdateValue(ctx, &pb.MatValue{Row: int32(0), Col: int32(0), Val: float32(0.5)})
			if err1 != nil{
				log.Fatalf("could not update value: %v", err1)
			}
			log.Printf("Status: %s", r1.Ret)
		}
	}
}