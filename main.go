// order_process project main.go
package main

import (
	"flag"
	"fmt"
	"order_process/process/db"
	"order_process/process/env"
	"order_process/process/service"
	"os"

	"github.com/Sirupsen/logrus"
)

var join string

func init() {
	flag.StringVar(&join, "join", "", "host:port of leader to join")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments]\n", os.Args[0])
		flag.PrintDefaults()
	}
}

// The entry of service
func main() {
	// Parse arguments
	flag.Parse()

	fmt.Println("Order Processing Service Start!")

	// Initialize environment
	env, err := env.New().InitEnv()
	if err != nil {
		logrus.Fatal(err)
		return
	}

	// Initialize database
	err = db.InitDatabase(env.RedisConfig.Host, env.RedisConfig.Port)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	// Create OrderProcessService instance and start.
	service := service.NewOrderProcessService(&env.ServiceConfig)
	err = service.Start(join)
	if err != nil {
		logrus.Fatal(err)
	}
}
