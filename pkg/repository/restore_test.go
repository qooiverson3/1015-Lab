package repository_test

import (
	"loader/pkg/domain"
	"loader/pkg/fake"
	"loader/pkg/repository"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func initial(m *mongo.Client) domain.Repository {
	return repository.NewLoaderRepo(m, &nats.Conn{}, fake.FakeStream)
}

func TestCreateTicket(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should none error and get _id", func(mt *mtest.T) {
		r := initial(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		ticket, err := r.CreateTicket(fake.FakeStream)

		assert.ErrorIs(t, err, nil)
		assert.NotEmpty(t, ticket)
	})
}

func TestGetTicket(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should none error and get restore log", func(mt *mtest.T) {
		r := initial(mt.Client)
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

		res, err := r.GetRestoreInfo()

		assert.Equal(t, fakeResult, res)
		assert.ErrorIs(t, nil, err)
	})
}

func TestDropCollection(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should none error", func(mt *mtest.T) {
		r := initial(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err := r.DropCollection()

		assert.ErrorIs(t, err, nil)
	})
}
