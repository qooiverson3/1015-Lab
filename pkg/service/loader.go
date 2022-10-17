package service

import (
	"context"
	"fmt"
	"loader/pkg/domain"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type LoaderService struct {
	Repo       domain.Repository
	StreamName string
}

func NewLoaderService(r *domain.Repository, stream string) *LoaderService {
	return &LoaderService{Repo: *r, StreamName: stream}
}

func (s LoaderService) Loader(stream string, ctx context.Context) {
	s.StreamName = stream

	js, err := s.Repo.GetJSCtx()
	if err != nil {
		log.Fatal("[ERR] get js context err:", err)
	}

	strInfo, err := js.StreamInfo(stream)
	if err != nil {
		log.Fatalf("[ERR] get stream info err:%v, stream:%v", err, stream)
	}

	js.DeleteConsumer(s.StreamName, fmt.Sprintf("%v-loader", s.StreamName))

	// get last sequence from mongodb
	mgoLastEvent, err := s.Repo.GetLastSeqIDFromDB(stream)
	if err != nil {
		log.Fatal("[ERR] get lastSeq from DB err, ", err)
	}

	log.Printf("[INF] Loader is runner.. (stream:%v startSequence:%v)\n", stream, mgoLastEvent.StreamSeq+1)

	// subscribe
	sub, err := js.PullSubscribe(">", fmt.Sprintf("%v-loader", s.StreamName), nats.BindStream(stream), nats.StartSequence(mgoLastEvent.StreamSeq+1), nats.MaxAckPending(1000))
	if err != nil {
		log.Fatalf("[ERR] %v (stream:%v)", err, stream)
	}

	go func() {
		for {
			msgs, _ := sub.Fetch(200)
			results := make([]interface{}, 200)

			if len(msgs) != 0 {
				for i := 0; i < len(msgs); i++ {
					meta, _ := msgs[i].Metadata()
					results[i] = domain.MessageInMgo{
						Message:    msgs[i],
						StreamSeq:  meta.Sequence.Stream,
						ReceivedAt: meta.Timestamp,
					}
					log.Printf("[INF] got seq:%v\n", meta.Sequence.Stream)
					msgs[i].Ack()
				}
				s.Repo.InsertMany(s.StreamName, results)
			}
			time.Sleep(time.Second)
		}
	}()

	// check seq state
	go s.checkStreamSequenceState(ctx, js, sub, strInfo, mgoLastEvent.MgoSeq)
}

func (s LoaderService) checkStreamSequenceState(ctx context.Context, js nats.JetStreamContext, sub *nats.Subscription, strInfo *nats.StreamInfo, mgoLastSeq uint64) {
	for {
		j, err := js.StreamInfo(s.StreamName)
		if err != nil {
			log.Fatalf("[ERR] can't get stream:%v info via goroutine, %v", s.StreamName, err)
		}

		c, err := sub.ConsumerInfo()
		if err != nil {
			log.Fatalf("[ERR] can't get consumer:%v info via goroutine, %v", c.Name, err)
		}

		select {
		case <-ctx.Done():
			log.Printf("[INF] received singal to unsubscribe %v..\n", c.Stream)
			time.Sleep(time.Second)
			sub.Unsubscribe()
			log.Printf("[INF] unsubscribe %v success\n", c.Stream)
			return
		default:
			// checking sequence status
			// when consumer DeliveredStreamSeq > stream lastSeq or mongodb lastSeq > stream lastSeq mean abnormal
			seqState := IsAbnormalSequenceState(c.Delivered.Stream, j.State.LastSeq, mgoLastSeq)

			log.Printf("[MONI] {\"strName\":\"%v\",\"conName\":\"%v\",\"con.str.seq\":\"%v\",\"str.last.seq\":\"%v\",\"state\":\"%v\"}", strInfo.Config.Name, c.Name, c.Delivered.Stream, j.State.LastSeq, seqState)
			time.Sleep(30 * time.Second)
		}
	}
}
