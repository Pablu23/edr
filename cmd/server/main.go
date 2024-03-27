package main

import (
	"edr/pkg/server"
	"go.uber.org/zap"
)

func main() {
	config := zap.NewDevelopmentConfig()
	//config.OutputPaths = append(config.OutputPaths, "log.log")
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	s := server.NewFromFile("running.txt", logger)
	s.ListenAndServe()
}
