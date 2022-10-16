package service

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func (s LoaderService) Follower(stream string) {
	js, err := s.Repo.GetJSCtx()
	if err != nil {
		log.Printf("[ERR] Elf get jetStreamContext fail,%v\n", err)
		return
	}

	_, err = js.StreamInfo(stream)
	if err != nil {
		log.Printf("[ERR] Elf get streamInfo fail,%v\n", err)
		return
	}

	log.Printf("[INF] The follower of %v is running..\n", stream)

	chMsgs := make(chan *nats.Msg, 4096)

	js.Subscribe(">", func(msg *nats.Msg) {
		chMsgs <- msg
	}, nats.BindStream(stream))

	go s.Store(chMsgs, "nats-follower", fmt.Sprintf("%v-%v", stream, time.Now()))
}
