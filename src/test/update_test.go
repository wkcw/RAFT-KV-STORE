package test

import (
	"fmt"
	"testing"
	"time"
)

func Test_Update_Matrix(t *testing.T) {
	var count = 0
	timer := time.After(20 * time.Second)
	for{
		select {
		case <-timer:
			fmt.Printf("Updated matrix for %d times\n", count)
			return
		default:
			parallel_clear()
			count++
		}
	}
}