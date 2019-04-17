package util

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func CreateConnMatrix(row int)  [][]float32 {
	var ret [][]float32

	for i := 0; i < row; i++ {
		tmp := make([]float32, row)
		ret = append(ret, tmp)
	}


	file, err := os.Open("/Users/wkcw/Desktop/cse223/garbage/cse223b-RAFT-KV-STORE/src/util/ConnMatrix.txt")

	if err != nil {
		log.Fatalf("could not find matrix file: %v", err)
	}

	defer file.Close()

	br := bufio.NewReader(file)

	for i := 0;i < row;i++ {
		line, _, err := br.ReadLine()

		if err != nil {
			if err != io.EOF {
				log.Fatalf("the format of matrix file is not corrected: %v", err)
			} else {
				log.Fatalf("unknown format: %v", err)

			}
			break
		}

		elements := strings.Split(string(line), " ")
		fmt.Println(len(elements))
		for j := 0;j < len(elements);j++ {
			t1, _ := strconv.ParseFloat(elements[j], 32)
			t2 := float32(t1)
			fmt.Println("Current j is %d" , j)
			ret[i][j] = t2
		}
	}

	return ret
}