// order_process project main.go
package main

import (
	"fmt"
	"log"
	"order_process/process/db"
	"order_process/process/env"
	"os"

	"github.com/Sirupsen/logrus"
)

// The entry of service
func main() {
	fmt.Println("Order Processing Service Start!")

	// Initialize logger
	logrus.SetOutput(os.Stdout)

	// Initialize environment
	env := env.New().InitEnv()

	// Initialize database
	err := db.InitDatabase(env.RedisConfig.Host, env.RedisConfig.Port)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Create OrderProcessService instance and start.
	orderProcessService := NewOrderProcessService()
	err = orderProcessService.Start()
	if err != nil {
		log.Fatal(err)
	}
}
