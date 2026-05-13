package main

import (
	"flag"
	"log"

	"github.com/joho/godotenv"
)


func init() {
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	log.Println("Environment loaded")
}
