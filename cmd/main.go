package main

import (
	"github.com/jiin-yang/messageBird/config"
	"github.com/jiin-yang/messageBird/internal/server"
	"github.com/rs/zerolog/log"
)

func main() {
	conf, err := config.New()
	checkFatalError(err)

	s := server.New(conf)

	err = s.Start()
	checkFatalError(err)
}

func checkFatalError(err error) {
	if err != nil {
		log.Fatal().Err(err)
	}
}
