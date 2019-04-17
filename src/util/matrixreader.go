package util

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func MatrixReader(row int, filename string)  [][]int {
	var ret [][]int

	for i := 0; i < row; i++ {
		tmp := make([]int, row)
		ret = append(ret, tmp)
	}

	file, err := os.Open(filename)
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
			tmp, _ := strconv.Atoi(elements[j])

			ret[i][j] = tmp
		}
	}

	return ret
}