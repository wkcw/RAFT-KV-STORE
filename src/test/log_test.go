package test

import (
	"client"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"math/rand"
	"path/filepath"
	pb "proto"
	"service"
	"strconv"
	"testing"
	"time"
	"util"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}


func Test_Log_1(t *testing.T) {
	config := util.CreateConfig()
	serverList := config.ServerList
	// generate random key, value pair
	var key string = RandStringRunes(10)
	var value string = RandStringRunes(10)
	fmt.Printf("Key generated is: %s\n", key)
	fmt.Printf("Value generated is: %s\n", value)

	// Set up a client to a set of servers
	client := client.NewClient(serverList)
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
			break
		}

		conn.Close()
		cancel()
	}

	time.Sleep(time.Millisecond * 5000)


	var pass bool = true

	for i := 0; i < serverList.ServerNum; i++ {
		filename := "STATE_CONFIG_" + strconv.Itoa(i) + ".json"
		path, _ := filepath.Abs("../util/" + filename)
		bytes, _ := ioutil.ReadFile(path)
		state := service.JsonToObject(bytes)
		entryList := state.GetLog().EntryList
		if len(entryList) == 0 {
			pass = false
			fmt.Printf("Server %d entryList is empty!!!\n", i)
			break
		}
		lastEntry := entryList[len(entryList) - 1]
		if lastEntry.GetOp() != "put" || lastEntry.GetKey() != key || lastEntry.GetVal() != value {
			pass = false
			fmt.Printf("Server %d does not have the log.\n", i)
			break
		}
		fmt.Printf("Server %d has the log.\n", i)
	}
	if pass {
		t.Log("Passed the log replication test.")
	} else {
		t.Error("Failed the log replication test.")
	}
}