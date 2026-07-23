package main

import (
	"log"
	"os"
	"time"

	"github.com/neomat-prog/cmd/server"
)

func main() {

	logger := log.New(os.Stdout, "", log.LstdFlags)

	srv := server.NewServer(server.ServerOpts{
		ListenAddr:   ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Logger:       logger,
	})

	if err := srv.Start(); err != nil {
		logger.Fatal(err)
	}

}
