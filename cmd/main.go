package main

import (
	"log"

	"github.com/Ka10kenHQ/watchclean.tv/internal/api"
	"github.com/Ka10kenHQ/watchclean.tv/internal/bootstrap"
)

func main() {
	bootstrap.Init()

	log.Println("Starting API server...")
	api.RunServer()
}

