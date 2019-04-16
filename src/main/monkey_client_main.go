package main

import (
	pb "chaosmonkey"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"time"
	"util"
)

func main() {
	serverlist := util.CreateServerList("/Users/cpwang/Desktop/cse223b-RAFT-KV-STORE/src/util/config.xml")
	// Set up a connection to the server.
	for {
		var monkey_operation, row, col, val, serverNum string
		serverNum = serverlist.ServerNum // total number of servers
		fmt.Scanln(&monkey_operation, &row, &col, &val)
		row_int, _ := strconv.ParseInt(row, 10, 32)
		col_int, _ := strconv.ParseInt(col, 10, 32)
		val_int, _ := strconv.ParseFloat(val, 32)
		serverNum_int, _ := strconv.ParseInt(serverNum, 10, 32)

		if (monkey_operation == "upload") {
			for _, server := range serverlist.Servers {
				var address string = server.Host + ":" + server.Port
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



				matrows := make([]pb.ConnMatrix_MatRow, serverNum_int)
				for i := 0; i < len(matrows); i++ {

					matrows[i].Vals = make([]float32, serverNum_int)
				}

				matrows_ptr := make([]*pb.ConnMatrix_MatRow, serverNum_int)
				for i := 0; i < len(matrows_ptr); i++ {
					matrows_ptr[i] = &matrows[i]
				}
				r, err := c.UploadMatrix(ctx, &pb.ConnMatrix{Rows: matrows_ptr})
				if err != nil {
					log.Fatalf("could not upload: %v", err)
				}
				log.Printf("Return code: %s", r.Ret)
			}
		}

		if (monkey_operation == "update") {
			for _, server := range serverlist.Servers{
				var address string = server.Host + ":" + server.Port
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
				r1, err1 := c.UpdateValue(ctx, &pb.MatValue{Row: int32(row_int), Col: int32(col_int), Val: float32(val_int)})
				if err1 != nil{
				log.Fatalf("could not get: %v", err1)
				}
				log.Printf("Status: %s", r1.Ret)
			}
		}
	}
}
