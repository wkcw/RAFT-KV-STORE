package main

import "service"

func main()  {
	kvService := service.NewKVService("localhost:9527")
	kvService.Start()
}