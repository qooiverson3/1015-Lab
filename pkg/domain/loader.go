package domain

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
)

type MessageInMgo struct {
	Message    *nats.Msg `bson:"message"`
	StreamSeq  uint64    `bson:"streamSequence"`
	ReceivedAt time.Time `bson:"receviedAt"`
}
type Chan struct {
	ChanService Service
	Stream      string
}

type Restore struct {
	ID        string    `bson:"_id"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdateAt  time.Time `bson:"updatedAt"`
	Have      int       `bson:"have"`
	Want      int       `bson:"want"`
	Stream    string    `bson:"stream"`
	Status    string    `bson:"status"`
	RestoreBy string    `bson:"restoreBt"`
	Message   string    `bson:"message"`
}

type Event struct {
	StreamSeq uint64    `bson:"streamSequence"`
	MgoSeq    uint64    `bson:"mgoSequence"`
	Data      *nats.Msg `bson:"message"`
}

type Repository interface {
	SaveToDB(msg *bson.D, db, coll string) error
	GetJSCtx() (nats.JetStreamContext, error)
	GetLastSeqIDFromDB(stream string) (Event, error)
	GetMsgFromMgoStore(opt string, filter Filter) []nats.Msg
	CreateTicket(stream string) (string, error)
	UpdateJobLogger(dr Restore)
	GetRestoreInfo() ([]Restore, error)
	IsRestoreInProgress(stream string) bool
	CloneCollection(stream string) bool
	DropCollection() error
	InsertMany(stream string, data []interface{}) bool
}

type Service interface {
	Loader(stream string, ctx context.Context)
	Restore(opt, uid string, filter Filter, stream string, cancelFunc context.CancelFunc, phase uint)
	GetTicket(stream string) (string, error)
	GetRestoreInfo() ([]Restore, error)
	IsRestoreInProgress(stream string) bool
	Follower(stream string)
}
