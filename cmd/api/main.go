package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	app "github.com/nevinmanoj/bhavana-backend/internal/app"
)

func main() {

	fmt.Println("Starting BHAVANA MANAGER API service...")

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}

	if err := app.Start(); err != nil {
		log.Fatal("Error starting server: ", err)
	}

}
