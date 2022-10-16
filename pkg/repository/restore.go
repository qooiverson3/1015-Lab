package repository

import (
	"context"
	"fmt"
	"loader/pkg/domain"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func scaleToEnd(opt string, filter domain.Filter) primitive.D {
	switch opt {
	case "seq":
		f := bson.D{
			{
				Key:   "$gte",
				Value: filter.Seq.Start,
			},
			{
				Key:   "$lte",
				Value: filter.Seq.End,
			},
		}

		if filter.Seq.End == 0 {
			f = bson.D{
				{
					Key: "$gte",
				},
			}
		}
		return bson.D{
			{
				Key:   "streamSequence",
				Value: f,
			},
		}
	case "time":
		var endTime time.Time = filter.Timeinterval.End
		if endTime == time.Time(time.Unix(0, 0)) {
			endTime = time.Now()
		}
		return bson.D{
			{
				Key: "receviedAt",
				Value: bson.D{
					{
						Key:   "$gte",
						Value: filter.Timeinterval.Start,
					},
					{
						Key:   "$lte",
						Value: endTime,
					},
				},
			},
		}
	}
	return bson.D{}
}

func (r *LoaderRepo) GetMsgFromMgoStore(opt string, filter domain.Filter) []nats.Msg {
	var messages []nats.Msg

	scale := scaleToEnd(opt, filter)

	cur, err := r.MongoCtx.MgoCollection.Find(context.TODO(), scale)
	if err != nil {
		log.Fatal("[ERR]", err)
	}

	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var message domain.Event
		cur.Decode(&message)
		messages = append(messages, *message.Data)
	}

	return messages
}

func (r *LoaderRepo) CreateTicket(stream string) (string, error) {
	createdAt := time.Now()
	col := r.MongoCtx.Client.Database("nats-log").Collection("restore")
	result, err := col.InsertOne(context.TODO(), bson.D{
		{
			Key: "createdAt", Value: time.Time(createdAt),
		},
		{
			Key: "status", Value: "Pending",
		},
		{
			Key: "stream", Value: stream,
		},
	})
	if err != nil {
		log.Printf("[WRN] fail to write log to mongodb before restore job")
		return "", err
	}

	id := result.InsertedID.(primitive.ObjectID)
	return id.Hex(), nil
}

func (r *LoaderRepo) UpdateJobLogger(dr domain.Restore) {
	updateAt := time.Now()
	id, err := primitive.ObjectIDFromHex(dr.ID)
	if err != nil {
		log.Printf("[ERR] failt ottocket to objectID,%v", err)
	}
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "status", Value: dr.Status},
		{Key: "updatedAt", Value: time.Time(updateAt)},
		{Key: "want", Value: dr.Want},
		{Key: "hava", Value: dr.Have},
		{Key: "message", Value: dr.Message},
		{Key: "restoreBy", Value: dr.RestoreBy},
	}}}

	col := r.MongoCtx.Client.Database("nats-log").Collection("restore")
	_, err = col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf("[WRN] fail to update ticket:%v,%v", dr.ID, err)
		return
	}
	log.Printf("[INF] restore stream sequence completed,ticket:%v", dr.ID)
}

func (r *LoaderRepo) GetRestoreInfo() ([]domain.Restore, error) {
	var result []domain.Restore
	col := r.MongoCtx.Client.Database("nats-log").Collection("restore")
	cur, err := col.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Println("[WRN] fail to get restore info")
	}

	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var e domain.Restore
		err = cur.Decode(&e)
		result = append(result, e)
	}
	return result, err
}

func (r *LoaderRepo) IsRestoreInProgress(stream string) bool {
	filter := bson.D{
		{Key: "$and", Value: bson.A{bson.D{
			{Key: "status", Value: bson.D{
				{Key: "$ne", Value: "Completed"},
			}}},
			bson.D{{Key: "stream", Value: bson.D{
				{Key: "$eq", Value: stream},
			}}},
		}},
	}
	col := r.MongoCtx.Client.Database("nats-log").Collection("restore")
	a, err := col.CountDocuments(context.TODO(), filter)
	if err != nil {
		log.Printf("[WRN] can't get %v log in restore table,%v", stream, err)
		return true
	}

	if a > 0 {
		return true
	}
	return false
}

func (r *LoaderRepo) CloneCollection(stream string) bool {
	sc := r.MongoCtx.MgoCollection
	amount, err := sc.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		log.Printf("[ERR] count docs fail,%v", err)
		return false
	}

	results := make([]interface{}, amount)

	cur, err := sc.Find(context.TODO(), &bson.D{})
	if err != nil {
		log.Printf("[ERR] get cursor fail,%v", err)
		return false
	}

	defer cur.Close(context.TODO())
	i := 0
	for cur.Next(context.Background()) {
		var data domain.MessageInMgo
		err := cur.Decode(&data)
		if err != nil {
			log.Printf("[ERR] get cursor next point fail,%v", err)
			return false
		}

		results[i] = data
		i++
	}

	dc := r.MongoCtx.Client.Database("nats-backup").Collection(fmt.Sprintf("%v-%v", stream, time.Now()))
	_, err = dc.InsertMany(context.TODO(), results)
	if err != nil {
		log.Printf("[ERR] backup fail,%v", err)
		return false
	}
	return true
}

func (r *LoaderRepo) DropCollection() error {
	col := r.MongoCtx.MgoCollection
	return col.Drop(context.TODO())
}
