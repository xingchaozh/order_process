// order_process project main.go
package main

import (
	"fmt"
	"log"
	"order_process/lib/db"
	"order_process/lib/model/pipeline"
	"os"

	"github.com/Sirupsen/logrus"
)

// The entry of service
func main() {
	fmt.Println("Order Processing Service Start!")

	// Initilize logger
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)

	// Initilize database
	db.InitDatabase("192.168.163.147", 6379)

	// Create OrderProcessService instance and start.
	orderProcessService := NewOrderProcessService(pipeline.NewProcessPipeline,
		pipeline.NewStepTaskHandler)
	err := orderProcessService.Start()
	if err != nil {
		log.Fatal(err)
	}
}
