package main

import (
	"time"

	"github.com/sirupsen/logrus"
)

const (
	templatePath = "/html-templates"
	initTime     = 10
)

func main() {

	logrus.SetFormatter(&logrus.JSONFormatter{})
	go scanServices()
	time.Sleep(initTime * time.Second)
	webserver()
}
