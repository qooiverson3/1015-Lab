package repository

import (
	"context"
	"loader/pkg/domain"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LoaderRepo struct {
	MongoCtx Mgo
	NATSCtx  Nats
}

type Mgo struct {
	Client        *mongo.Client
	MgoCollection *mongo.Collection
}

type Nats struct {
	Conn   *nats.Conn
	Stream string
}

func NewLoaderRepo(db *mongo.Client, nc *nats.Conn, stream string) domain.Repository {
	collection := db.Database("nats").Collection(stream)
	return &LoaderRepo{
		MongoCtx: Mgo{
			Client:        db,
			MgoCollection: collection,
		},
		NATSCtx: Nats{
			Conn:   nc,
			Stream: stream,
		},
	}
}

func (r *LoaderRepo) SaveToDB(msg *bson.D, db, coll string) error {
	col := r.MongoCtx.Client.Database(db).Collection(coll)
	_, err := col.InsertOne(context.TODO(), msg)
	//_, err := r.MongoCtx.MgoCollection.InsertOne(context.TODO(), msg)
	if err != nil {
		return err
	}
	return nil
}

func (r *LoaderRepo) GetJSCtx() (nats.JetStreamContext, error) {
	return r.NATSCtx.Conn.JetStream()
}

func (r *LoaderRepo) GetLastSeqIDFromDB(stream string) (domain.Event, error) {
	var e domain.Event
	opts := options.Find()

	opts.SetSort(bson.D{{Key: "streamSequence", Value: -1}}).SetLimit(1)
	cursor, err := r.MongoCtx.MgoCollection.Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		return domain.Event{}, err
	}

	for cursor.Next(context.TODO()) {
		err = cursor.Decode(&e)
		return e, err
	}
	return domain.Event{}, nil
}
