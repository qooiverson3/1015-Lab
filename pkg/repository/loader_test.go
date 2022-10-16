package repository_test

import (
	"loader/pkg/fake"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestSaveToDB(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		r := initial(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err := r.SaveToDB(&bson.D{}, "nats", fake.FakeStream)

		assert.ErrorIs(t, nil, err)
	})
}

func TestGetJSCtx(t *testing.T) {
	r := initial(&mongo.Client{})

	nc := &nats.Conn{}
	j, ferr := nc.JetStream()
	js, err := r.GetJSCtx()

	assert.Equal(t, j, js)
	assert.ErrorIs(t, ferr, err)
}
