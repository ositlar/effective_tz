package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"tz/internal/apiserver"
	"tz/internal/config"
)

func main() {
	cfg := config.MustConfig()
	if err := apiserver.Start(*cfg); err != nil {
		log.Fatalln(err)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	fmt.Println("Application stopped")
}
