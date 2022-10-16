package service

import (
	"context"
	"fmt"
	"loader/pkg/domain"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func (s LoaderService) Restore(opt, ticket string, optFilter domain.Filter, stream string, cancelFunc context.CancelFunc, phase uint) {
	messages := s.Repo.GetMsgFromMgoStore(opt, optFilter)
	var want int = len(messages)
	var have int

	result := domain.Restore{
		ID:        ticket,
		Have:      have,
		Want:      want,
		RestoreBy: opt,
		Message:   "",
	}

	js, err := s.Repo.GetJSCtx()
	if err != nil {
		log.Printf("[ERR] %v", err)
	}

	result.Status = "In Progress"
	s.Repo.UpdateJobLogger(result)

	log.Printf("[INF] start restore..\n")

	for i := 0; i < want; i++ {
		meta, err := messages[i].Metadata()

		if err != nil {
			log.Printf("[ERR] %v\n", err)
			result.Status = "Failure"
			result.Message = "can't get nats.msg metadata (sequence:%v)"
			s.Repo.UpdateJobLogger(result)
			return
		}

		if s.IsPubSuccessful(&messages[i], uint64(i)+meta.Sequence.Stream, stream, js, ticket) {
			have++
		}
		time.Sleep(5 * time.Millisecond)
	}

	// backup collection to nats-backup database and reset collection
	log.Printf("[INF] start to backup %v collection..\n", stream)
	if !s.Backup(stream) {
		result.Status = "Failure"
		result.Message = fmt.Sprintf("Backup %v collection fail", stream)
		s.Repo.UpdateJobLogger(result)

		log.Printf("[INF] backup %v collection completed\n", stream)
		return
	}
	log.Printf("[INF] backup %v collection completed\n", stream)

	switch phase {
	case 1:
		time.Sleep(time.Second)
		cancelFunc()
		log.Printf("[INF] start to unsubscirbe..\n")
		time.Sleep(30 * time.Second)
		log.Printf("[INF] unsubscribe completed\n")
		// drop
		log.Printf("[INF] start to drop %v collection..\n", stream)

		if s.StreamName == stream {
			if err = s.Drop(); err != nil {
				result.Status = "Failure"
				result.Message = fmt.Sprintf("Drop %v collection fail", stream)
				s.Repo.UpdateJobLogger(result)

				log.Printf("[INF] drop %v collection fail\n", stream)
				return
			}
			log.Printf("[INF] drop %v collection completed\n", stream)
		}

		// jobs success
		result.Status = "Completed"
		result.Message = "All jobs completed"
		result.Have = have
		s.Repo.UpdateJobLogger(result)

		log.Fatal("[INF] restore completed and exit process")
	default:
		result.Status = "Completed"
		result.Message = "All jobs completed"
		result.Have = have
		s.Repo.UpdateJobLogger(result)
		log.Printf("[INF] restore completed.(ticket:%v\n", ticket)
	}
}

func (s LoaderService) IsPubSuccessful(msg *nats.Msg, seq uint64, stream string, js nats.JetStreamContext, ticket string) bool {

	msg.Reply = ""

	// add reload header
	if msg.Header == nil {
		msg.Header = nats.Header{}
		msg.Header["X-RESTORE"] = []string{fmt.Sprintf("%v", time.Now().Format(time.RFC3339)), ticket}
	} else {
		msg.Header.Add("X-RESTORE", fmt.Sprintf("%v", time.Now().Format(time.RFC3339)))
		msg.Header.Add("X-RESTORE", ticket)
	}

	_, err := js.PublishMsgAsync(msg)
	if err != nil {
		log.Printf("[ERR] publish fail, %v", err)
		return false
	}

	select {
	case <-js.PublishAsyncComplete():
		return true
	case <-time.After(5 * time.Second):
		log.Printf("[WRN] restore sequence:%v to stream:%v", seq, stream)
	}
	return false
}

func (s LoaderService) GetTicket(stream string) (string, error) {
	return s.Repo.CreateTicket(stream)
}

func (s LoaderService) GetRestoreInfo() ([]domain.Restore, error) {
	return s.Repo.GetRestoreInfo()
}

func (s LoaderService) IsRestoreInProgress(stream string) bool {
	return s.Repo.IsRestoreInProgress(stream)
}

func (s LoaderService) Backup(stream string) bool {
	return s.Repo.CloneCollection(stream)
}

func (s LoaderService) Drop() error {
	return s.Repo.DropCollection()
}
