package main

import (
	"client"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	pb "proto"
	"time"
	"util"
)

var (
	operation, key, value string
)

func main() {
	config := util.CreateConfig()
	serverList := config.ServerList
	// Set up a client to a set of servers
	client := client.NewClient(serverList)
	sequenceNo := int64(0)
	for {
		fmt.Scanln(&operation, &key, &value)
		//operation = "put"
		if operation == "put" {
			var address string
			address = client.PickRandomServer()
			sequenceNo++
			for {
				// connect to the server
				conn, err := grpc.Dial(address, grpc.WithInsecure())
				if err != nil {
					log.Fatalf("did not connect: %v", err)
				}
				c := pb.NewKeyValueStoreClient(conn)

				// Contact the server and print out its response.
				ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)

				response, errCode := c.Put(ctx, &pb.PutRequest{Key: key, Value: value})
				if errCode != nil {
					log.Printf("could not put raft, an timeout occurred: %v", errCode)
					address = client.PickRandomServer()
					continue
				}


				if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
					// if the return address is not leader
					leaderID := response.LeaderID
					if leaderID == -1{
						address = client.PickRandomServer()
					}else{
						leaderServer := client.ServerList.Servers[leaderID]
						address = leaderServer.Addr
					}

					conn.Close()
					cancel()
					continue;
				}

				if response.Ret == pb.ReturnCode_SUCCESS {
					log.Println("Put to the leader successfully")

					conn.Close()
					cancel()
					break;
				}

				conn.Close()
				cancel()
			}
		}

		//operation = "put block"
		if operation == "putb" {
			var address string
			address = client.PickRandomServer()
			sequenceNo++
			//numVal, valErr := strconv.Atoi(value)
			//if valErr != nil{
			//	log.Printf("please input a number")
			//}
			block := "a"
			//block_unit := "aaaaaaaaaaaaaaaa"
			for i:=0; i<24; i++{
				block += block
			}

			fmt.Printf("%v start", time.Now())
			for {
				// connect to the server
				conn, err := grpc.Dial(address, grpc.WithInsecure())
				if err != nil {
					log.Fatalf("did not connect: %v", err)
				}
				c := pb.NewKeyValueStoreClient(conn)

				// Contact the server and print out its response.
				ctx, cancel := context.WithTimeout(context.Background(), 50 * time.Second)

				response, errCode := c.Put(ctx, &pb.PutRequest{Key: key, Value: block})
				if errCode != nil {
					log.Printf("could not put raft, an timeout occurred: %v", errCode)
				}


				if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
					// if the return address is not leader
					leaderID := response.LeaderID
					leaderServer := client.ServerList.Servers[leaderID]
					address = leaderServer.Addr

					conn.Close()
					cancel()
					continue;
				}

				if response.Ret == pb.ReturnCode_SUCCESS {
					log.Println("Put to the leader successfully")

					conn.Close()
					cancel()
					break;
				}

				conn.Close()
				cancel()
			}

			fmt.Printf("%v end", time.Now())
		}

		if operation == "get" || operation == "getb"{
			var address string
			address = client.PickRandomServer()
			for {
				// connect to the server
				conn, err := grpc.Dial(address, grpc.WithInsecure())
				if err != nil {
					log.Fatalf("did not connect: %v", err)
				}
				c := pb.NewKeyValueStoreClient(conn)

				// Contact the server and print out its response.
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

				response, errCode := c.Get(ctx, &pb.GetRequest{Key: key})
				if errCode != nil {
					log.Printf("could not get raft, an timeout occurred: %v", errCode)
				}

				if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
					// if the return address is not leader
					leaderID := response.LeaderID
					leaderServer := client.ServerList.Servers[leaderID]
					address = leaderServer.Addr

					conn.Close()
					cancel()
					continue;
				}

				if response.Ret == pb.ReturnCode_FAILURE_GET_NOKEY {
					log.Printf("No such Key")
				}

				if response.Ret == pb.ReturnCode_SUCCESS {
					value := response.Value
					if operation == "get"{
						log.Printf("Get the key successfully %s, the value is: %s", key,  value)
					}
					if operation == "getb"{
						log.Printf("Get the key successfully %s, the big block length is: %s", key,  len(value))
					}
					conn.Close()
					cancel()
					break;
				}

				conn.Close()
				cancel()
			}
		}
	}
}