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
					log.Fatalf("could not update value: %v", err1)
				}
				log.Printf("Status: %s", r1.Ret)
			}
		}

		if (monkey_operation == "kill") {
			serverID := &row
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

				r1, err1 := c.KillServer(ctx, &pb.ServerStat{ServerID:serverID})
				if err1 != nil{
					log.Fatalf("could not kill the server: %v", err1)
				}
				log.Printf("Status: %s", r1.Ret)
			}
		}


		if (monkey_operation == "partition") {

			fmt.Printf("Please enter the server ids for a single partition:")
			reader := bufio.NewReader(os.Stdin)
			Input, _, _ := reader.ReadLine()
			var servers string
			servers = string(Input)
			fmt.Println(servers)
			s := strings.Split(servers, " ")
			idMap := make(map[int]string)
			for _, id := range s {
				idInt, _ := strconv.ParseInt(id, 10, 32)
				idMap[int(idInt)] = id
			}
			for _, server := range serverlist.Servers{
				//conncet to every server
				id := server.ServerId
				address := server.Host + ":" + server.Port
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
				// if server id in the map, connect server id with every other server in the map and cut connection
				// with any other servers that are not in the map
				_, ok := idMap[id]
				if (ok) {
					for row := 0; row < serverNum; row++ {
						_, ok_sender := idMap[row]
						if (ok_sender) {
							c.UpdateValue(ctx, &pb.MatValue{Row: int32(row), Col: int32(id), Val: float32(0)})
						} else {
							c.UpdateValue(ctx, &pb.MatValue{Row: int32(row), Col: int32(id), Val: float32(1)})
						}
					}

					for col := 0; col < serverNum; col++ {
						_, ok_receiver := idMap[col]
						if (ok_receiver) {
							c.UpdateValue(ctx, &pb.MatValue{Row: int32(id), Col: int32(col), Val: float32(0)})
						} else {
							c.UpdateValue(ctx, &pb.MatValue{Row: int32(id), Col: int32(col), Val: float32(0)})
						}
					}
				} else { // if server id is not in the map, cut connections with the servers that are in the map
					for row := 0; row < serverNum; row++ {
						_, ok_sender := idMap[row]
						if (ok_sender) {
							c.UpdateValue(ctx, &pb.MatValue{Row: int32(row), Col: int32(id), Val: float32(1)})
						}
					}

					for col:= 0; col < serverNum; col++ {
						_, ok_receiver := idMap[col]
						if (ok_receiver) {
							c.UpdateValue(ctx, &pb.MatValue{Row: int32(id), Col: int32(col), Val: float32(1)})
						}
					}
				}
			}

		}
	}
}