package main

import (
	"time"
	"fmt"
)

func su(){

}

func sub(){
	ticker := time.NewTicker(time.Second)
	for range ticker.C{
		fmt.Println("999")
	}
}

func main(){
	done := make(chan bool)
	for i:=0; i<10; i++{
		go sub()
	}
	<- done
}