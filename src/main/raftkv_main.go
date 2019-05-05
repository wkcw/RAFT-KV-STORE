package main

import (
	"os"
	"service"
)

func main()  {
	ID := os.Args[1]
	kvService := service.NewKVService(ID)
	kvService.Start()
}