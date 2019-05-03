package main

import (
	"bufio"
	pb "chaosmonkey"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"util"
)

func main() {
	config := util.CreateConfig()
	serverlist := config.ServerList
	// Set up a connection to the server.
	for {
		var monkey_operation, row, col, val string
		var serverNum int
		serverNum = serverlist.ServerNum // total number of servers
		fmt.Scanln(&monkey_operation, &row, &col, &val)
		row_int, _ := strconv.ParseInt(row, 10, 32)
		col_int, _ := strconv.ParseInt(col, 10, 32)
		val_int, _ := strconv.ParseFloat(val, 32)
		if (monkey_operation == "upload") {
			for _, server := range serverlist.Servers {
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
				matrix := util.CreateConnMatrix(serverNum)
				matrows := make([]pb.ConnMatrix_MatRow, serverNum)
				for i := 0; i < len(matrows); i++ {

					matrows[i].Vals = matrix[i]
				}

				matrows_ptr := make([]*pb.ConnMatrix_MatRow, serverNum)
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
				r1, err1 := c.UpdateValue(ctx, &pb.MatValue{Row: int32(row_int), Col: int32(col_int), Val: float32(val_int)})
				if err1 != nil{
					log.Fatalf("could not update value: %v", err1)
				}
				log.Printf("Status: %s", r1.Ret)
			}
		}

		if (monkey_operation == "kill") {
			serverID := row
			serverID_int, _ := strconv.ParseInt(serverID, 10, 32)
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

				r1, err1 := c.KillServer(ctx, &pb.ServerStat{ServerID:int32(serverID_int)})
				if err1 != nil{
					log.Fatalf("could not kill the server: %v", err1)
				}
				log.Printf("Status: %s", r1.Ret)
			}
		}


		if (monkey_operation == "partition") {

			fmt.Println("Please enter the server ids for a single partition:")
			reader := bufio.NewReader(os.Stdin)
			Input, _, _ := reader.ReadLine()

			var servers []*pb.Server
			s := strings.Split(string(Input), " ")

			for _, id := range s {
				serverID, _ := strconv.ParseInt(id, 10, 32)
				var server pb.Server
				server.ServerID = int32(serverID)
				servers = append(servers, &server)
				fmt.Println(server.ServerID)
			}

			for _, server := range serverlist.Servers{
				// connect to every server
				address := server.Addr
				// connect to the server
				conn, err := grpc.Dial(address, grpc.WithInsecure())
				if err != nil {
					log.Fatalf("did not connect: %v", err)
				}
				defer conn.Close()
				c := pb.NewChaosMonkeyClient(conn)
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				c.Partition(ctx, &pb.PartitionInfo{Server: servers})
				// Contact the server and print out its response.
				defer cancel()
			}

		}
	}
}