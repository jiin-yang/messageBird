package main

import (
	"fmt"
	"github.com/jiin-yang/messageBird/config"
	"github.com/jiin-yang/messageBird/internal/server"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	conf, err := config.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Config error:", err)
		os.Exit(1)
	}

	s := server.New(conf)

	if err = s.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server start failed")
	}
}
