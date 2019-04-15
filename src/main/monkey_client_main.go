package main

import (
	pb "chaosmonkey"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"time"
)

const (
	address     = "127.0.0.1:9528"
	defaultName = "world"
)
var (
	operation, row, col, val string
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewChaosMonkeyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for {
		fmt.Scanln(&operation, &row, &col, &val)
		row, _ := strconv.ParseInt(row, 10, 32)
		col, _ := strconv.ParseInt(col, 10, 32)
		val, _ := strconv.ParseFloat(val, 32)
		
		if (operation == "upload") {
			matrows := make([]pb.ConnMatrix_MatRow, 5)
			for i := 0; i < 5; i++ {
				
				matrows[i].Vals = []float32 {0.5, 0.5, 0.5, 0.5, 0.5}
			}
			matrows_ptr := make([]*pb.ConnMatrix_MatRow, 5)
			for i := 0; i < 5; i++ {
				matrows_ptr[i] = &matrows[i]
			}
			r, err := c.UploadMatrix(ctx, &pb.ConnMatrix{Rows: matrows_ptr})
			if err != nil {
				log.Fatalf("could not upload: %v", err)
			}
			log.Printf("Return code: %s", r.Ret)
		}

		if (operation == "update") {
			r1, err1 := c.UpdateValue(ctx, &pb.MatValue{Row: int32(row), Col: int32(col), Val: float32(val)})
			if err1 != nil {
				log.Fatalf("could not get: %v", err1)
			}
			log.Printf("Status: %s", r1.Ret)
		}
	}
}
