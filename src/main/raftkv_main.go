package main

import "service"

func main()  {
	kvService := service.NewKVService("localhost:10528")
	kvService.Start()
}