package service_test

import (
	"context"
	"loader/pkg/domain"
	"loader/pkg/fake"
	"loader/pkg/repository"
	"loader/pkg/service"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func initial(m *mongo.Client) *service.LoaderService {
	r := repository.NewLoaderRepo(m, &nats.Conn{}, fake.FakeStream)
	return service.NewLoaderService(&r, fake.FakeStream)
}

func TestGetTicket(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should none error and get _id", func(mt *mtest.T) {
		s := initial(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		ticket, _ := s.GetTicket(fake.FakeStream)
		id, err := primitive.ObjectIDFromHex(ticket)

		col := mt.Client.Database("nats-log").Collection("restore")
		col.DeleteOne(context.TODO(), bson.M{"_id": id})

		assert.ErrorIs(t, err, nil)
		assert.NotEmpty(t, ticket)
	})
}

func TestGetRestoreInfo(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should no-error and get restore log", func(mt *mtest.T) {
		s := initial(mt.Client)
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "nats-log.restore", mtest.FirstBatch, bson.D{
				{Key: "stream", Value: fake.FakeStream},
				{Key: "Status", Value: "Completed"},
			}),
			mtest.CreateCursorResponse(2, "nats-log.restore", mtest.NextBatch, bson.D{
				{Key: "stream", Value: fake.FakeStream},
				{Key: "Status", Value: "In Progress"},
			}))

		fakeData1 := domain.Restore{
			Stream: fake.FakeStream,
			Status: "Completed",
		}

		fakeData2 := domain.Restore{
			Stream: fake.FakeStream,
			Status: "In Progress",
		}

		fakeResult := []domain.Restore{}
		fakeResult = append(fakeResult, fakeData1, fakeData2)

		r, err := s.GetRestoreInfo()

		assert.Equal(t, fakeResult, r)
		assert.ErrorIs(t, nil, err)
	})
}

/*
func TestIsRestoreInProgress(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should return false", func(mt *mtest.T) {
		s := initial(mt.Client)

		filter := bson.D{
			{Key: "$and", Value: bson.A{bson.D{
				{Key: "status", Value: bson.D{
					{Key: "$ne", Value: "Completed"},
				}}},
				bson.D{{Key: "stream", Value: bson.D{
					{Key: "$eq", Value: fake.FakeStream},
				}}},
			}},
		}
		mt.Coll.CountDocuments(context.TODO(), filter)
		//mt.AddMockResponses(bson.D{})

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "nats-log.restore",mtest))

		result := s.IsRestoreInProgress(fake.FakeStream)
		assert.Equal(t, result, false)
	})
}
*/

func TestDrop(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should none error and get _id", func(mt *mtest.T) {
		s := initial(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err := s.Drop()
		assert.ErrorIs(t, err, nil)
	})
}
