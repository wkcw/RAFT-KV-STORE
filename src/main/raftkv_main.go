package main

import "service"

func main()  {
	kvService := service.NewKVService()
	kvService.Start()
}