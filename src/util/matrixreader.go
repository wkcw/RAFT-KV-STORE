package util

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func CreateConnMatrix(row int)  [][]float32 {
	var ret [][]float32

	for i := 0; i < row; i++ {
		tmp := make([]float32, row)
		ret = append(ret, tmp)
	}

	path, _ := filepath.Abs("../util/ConnMatrix.txt")
	file, err := os.Open(path)

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
		for j := 0;j < len(elements);j++ {
			t1, _ := strconv.ParseFloat(elements[j], 32)
			t2 := float32(t1)
			ret[i][j] = t2
		}
	}

	return ret
}
