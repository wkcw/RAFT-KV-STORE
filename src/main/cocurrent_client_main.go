package main

import (
	"client"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"os"
	"path/filepath"
	pb "proto"
	"strconv"
	"time"
	"util"
)

func main(){
	clientNum := os.Args[1]
	intClientNum, _ := strconv.Atoi(clientNum)


	for uniqueClientId:=0; uniqueClientId<intClientNum; uniqueClientId++{
		go clientRoutine(uniqueClientId, intClientNum)
	}
	pause := make(chan bool)
	pause <- true
}

func clientRoutine(startNo int, incStep int)  {
	logDir := "concurrent_client_log"
	checkPathExistenceAndCreate(logDir)
	logPath, _ := filepath.Abs("./"+logDir+"/")
	fileName := fmt.Sprintf("concurrent_client_%d.log", startNo)
	logFile,err  := os.OpenFile(logPath+"/"+fileName, os.O_RDWR|os.O_CREATE, 0666)
	defer logFile.Close()
	if err != nil {
		log.Fatalln("open file error ! %v", err)
	}
	// 创建一个日志对象
	debugLog := log.New(logFile,"[Debug]",log.LstdFlags)


	config := util.CreateConfig()
	serverList := config.ServerList
	// Set up a client to a set of servers
	client := client.NewClient(serverList)
	sequenceNo := int64(startNo)

	incKey := 0
	incVal := 0


	for {
		incKey++
		incVal++
		fmt.Printf("%v\n", time.Now())
		//operation == "put"
		var address string
		address = client.PickRandomServer()
		sequenceNo += int64(incStep)
		for {
			// connect to the server
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				debugLog.Fatalf("did not connect: %v", err)
			}
			c := pb.NewKeyValueStoreClient(conn)

			// Contact the server and print out its response.
			ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)

			strkey := fmt.Sprintf("client%d", startNo) + strconv.Itoa(incKey)
			strval := strkey

			response, errCode := c.Put(ctx, &pb.PutRequest{Key: strkey, Value: strval})
			if errCode != nil {
				debugLog.Printf("could not put raft, an timeout occurred: %v", errCode)
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
				debugLog.Println("Put to the leader successfully")

				conn.Close()
				cancel()
				break;
			}

			conn.Close()
			cancel()
		}
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkPathExistenceAndCreate(_dir string){
	exist, err := PathExists(_dir)
	if err != nil {
		fmt.Printf("get dir error![%v]\n", err)
		return
	}

	if exist {
	} else {
		// 创建文件夹
		err := os.Mkdir(_dir, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
		}
	}
}