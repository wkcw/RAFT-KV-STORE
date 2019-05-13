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
	durationStr := os.Args[2]
	intClientNum, _ := strconv.Atoi(clientNum)
	durationInt, _ := strconv.Atoi(durationStr)
	durantion := time.Duration(durationInt)
	leaderAddress := sniffLeader()
	fmt.Printf("leaderAddress: %s\n", leaderAddress)

	resultChan := make(chan int, intClientNum)
	for uniqueClientId:=0; uniqueClientId<intClientNum; uniqueClientId++{
		go clientRoutine(leaderAddress, uniqueClientId, intClientNum, durantion*time.Second, resultChan)
	}
	myResult := 0
	for i:=0; i<intClientNum; i++{
		myResult += <- resultChan
	}
	fmt.Printf("myResult: %d", myResult)
}

func clientRoutine(address string, startNo int, incStep int, duration time.Duration, resultChan chan int) int {
	logDir := "concurrent_client_log"
	checkPathExistenceAndCreate(logDir)
	logPath, _ := filepath.Abs("./"+logDir+"/")
	fileName := fmt.Sprintf("concurrent_client_%d.log", startNo)
	logFile,err  := os.OpenFile(logPath+"/"+fileName, os.O_RDWR|os.O_CREATE, 0666)
	defer logFile.Close()
	if err != nil {
		log.Println("open file error ! %v", err)
	}
	// 创建一个日志对象
	debugLog := log.New(logFile,"[Debug]",log.LstdFlags)


	// connect to the server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		debugLog.Printf("did not connect: %v", err)
	}
	c := pb.NewKeyValueStoreClient(conn)
	timer := time.After(duration)
	succCnt := 0
	for {
		select{
		case <- timer:
			resultChan <- succCnt
		default:
			// Contact the server and print out its response.
			ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)

			response, errCode := c.Put(ctx, &pb.PutRequest{Key: "concurrent", Value: "put"})
			if errCode != nil {
				debugLog.Printf("could not put raft, an timeout occurred: %v", errCode)
			}
			if response.Ret == pb.ReturnCode_SUCCESS {
				succCnt++
				cancel()
				continue
			}
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

func sniffLeader() string{
	config := util.CreateConfig()
	serverList := config.ServerList
	// Set up a client to a set of servers
	client := client.NewClient(serverList)
	address := client.PickRandomServer()
	for {
		// connect to the server
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Printf("did not connect: %v", err)
		}
		c := pb.NewKeyValueStoreClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)

		response, errCode := c.Put(ctx, &pb.PutRequest{Key: "sniff", Value: "leader"})
		if errCode != nil {
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
			return address
		}

		conn.Close()
		cancel()
	}
}