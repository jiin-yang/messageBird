package message

import (
	"context"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	cronJobFrequencySeconds = 10
)

type Cron struct {
	messageUseCase UseCase
	StopChan       chan bool
	IsRunning      bool
	cancelFunc     context.CancelFunc
}

func NewCron(messageUseCase UseCase) *Cron {
	return &Cron{
		messageUseCase: messageUseCase,
		StopChan:       make(chan bool),
		IsRunning:      false,
	}
}

func (c *Cron) StartCron() {
	if c.IsRunning {
		log.Warn().Msg("Cron job is already running - cron.StartCron")
		return
	}
	c.IsRunning = true

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel

	log.Info().Msg("Cron job started - cron.StartCron")
	go func() {
		for {
			select {
			case <-c.StopChan:
				log.Info().Msg("Cron job stopped - cron.StartCron")
				c.IsRunning = false
				return
			default:
				log.Info().Msgf("Executing cron job - cron.StartCron")
				err := c.messageUseCase.SendMessages(ctx)
				if err != nil {
					log.Error().Err(err).Msg("Error executing cron job - cron.StartCron")
				}
				time.Sleep(cronJobFrequencySeconds * time.Second)
			}
		}
	}()
}

func (c *Cron) StopCron() {
	if !c.IsRunning {
		log.Warn().Msg("Cron job is not running - cron.StopCron")
		return
	}
	c.StopChan <- true
	if c.cancelFunc != nil {
		c.cancelFunc()
		log.Info().Msg("Context canceled - cron.StopCron")
	}
}
