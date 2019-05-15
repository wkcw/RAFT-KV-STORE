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
	config := util.CreateConfig()
	clientNum := os.Args[1]
	durationStr := os.Args[2]
	intClientNum, _ := strconv.Atoi(clientNum)
	durationInt, _ := strconv.Atoi(durationStr)
	durantion := time.Duration(durationInt)
	leaderAddress := sniffLeader()
	leaderID := -1
	for i, serverdes := range config.ServerList.Servers{
		if serverdes.Addr==leaderAddress{
			leaderID = i
		}
	}
	fmt.Printf("leaderAddress: %s\n", leaderAddress)

	resultChan := make(chan int, intClientNum)
	//killLeader(config)
	for uniqueClientId:=0; uniqueClientId<intClientNum; uniqueClientId++{
		go clientRoutine(leaderAddress, uniqueClientId, intClientNum, durantion*time.Second, resultChan, int32(leaderID))
	}
	myResult := 0
	for i:=0; i<intClientNum; i++{
		myResult += <- resultChan
	}
	fmt.Printf("myResult: %d", myResult)
}

func clientRoutine(addressArg string, startNo int, incStep int, duration time.Duration, resultChan chan int, leaderID int32) int {
	key := "concurrent"
	value := generate_big_value(1)

	config := util.CreateConfig()
	serverList := config.ServerList
	client := client.NewClient(serverList)
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

	address := addressArg
	// connect to the server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		debugLog.Printf("did not connect: %v", err)
	}
	c := pb.NewKeyValueStoreClient(conn)
	timer := time.After(duration)
	succCnt := 0
	//quit := make(chan bool)
	for {
		select{
		case <- timer:
			resultChan <- succCnt
		default:
			// Contact the server and print out its response.
			ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)

			response, errCode := c.Put(ctx, &pb.PutRequest{Key: key, Value: value})
			if response == nil {
				conn.Close()
				cancel()
				for{
					address = client.PickRandomServer()
					if address != serverList.Servers[leaderID].Addr{
						break
					}
				}
				conn, err := grpc.Dial(address, grpc.WithInsecure())
				if err != nil {
					debugLog.Printf("did not connect: %v", err)
				}
				c = pb.NewKeyValueStoreClient(conn)
				continue
			}
			if errCode != nil {
				debugLog.Printf("could not put raft, an timeout occurred: %v", errCode)
				continue
			}
			if response.Ret == pb.ReturnCode_SUCCESS {
				succCnt++
				cancel()
				continue
			}
			if response.Ret == pb.ReturnCode_FAILURE_GET_NOTLEADER {
				reconnLeaderID := response.LeaderID
				// if the return address is not leader
				if reconnLeaderID == leaderID || reconnLeaderID == -1 {
					for{
						address = client.PickRandomServer()
						if address != serverList.Servers[leaderID].Addr{
							break
						}
					}
				} else {
					address = serverList.Servers[reconnLeaderID].Addr
				}
				conn.Close()
				cancel()
				// connect to the server
				conn, err := grpc.Dial(address, grpc.WithInsecure())
				if err != nil {
					debugLog.Printf("did not connect: %v", err)
				}
				c = pb.NewKeyValueStoreClient(conn)
				continue;
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
				log.Println("Chosen random server: "+address)
			}else{
				leaderServer := client.ServerList.Servers[leaderID]
				address = leaderServer.Addr
				log.Println("Chosen random server: "+address)
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

func sniffLeaderWithQuit(quit chan bool) string{
	config := util.CreateConfig()
	serverList := config.ServerList
	// Set up a client to a set of servers
	client := client.NewClient(serverList)
	address := client.PickRandomServer()
	for {
		select{
		case <-quit:
			return ""
		default:
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
					log.Println("Chosen random server: "+address)
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
}


func generate_big_value(order int) string{
	value := "a"
	for i:=0; i<order; i++{
		value += value
	}
	return value
}


func killLeader(config *util.Config) {
	serverlist := config.ServerList
	// check partitioned group
	for _, server := range serverlist.Servers {

		address := serverlist.Servers[server.ServerId].Addr
		// connect to the server
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		defer conn.Close()
		client := pb.NewKeyValueStoreClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		// generating random params for
		response, _ := client.IsLeader(ctx, &pb.CLRequest{})
		if response != nil && response.IsLeader == true {
			response, _ := client.Exit(ctx, &pb.ExitRequest{})

			if response == nil || response.Ret == pb.ReturnCode_SUCCESS {
				break
			}

		}
	}

}