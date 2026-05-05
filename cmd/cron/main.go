package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/jasonbronson/go-gin-boilerplate/config"

	"github.com/robfig/cron/v3"
)

func main() {

	config.Cfg.NewRelicApp = config.NewRelicApp()
	log.Println("=====cron job ======")
	c := cron.New()
	// Add the jobs here and please keep the consistency of naming convention, filename in snake case, and interval and job function in camel case
	//c.AddFunc(jobs.DoSomething, jobs.DoSomething)
	c.Start()
	log.Println("=====cron system started======")

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

}
