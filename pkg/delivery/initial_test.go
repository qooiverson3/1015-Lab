package delivery_test

import (
	"context"
	"loader/pkg/delivery"
	"loader/pkg/domain"
	"loader/pkg/fake"
	"loader/pkg/repository"
	"loader/pkg/service"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func initial(m *mongo.Client) (chan domain.Chan, context.CancelFunc) {
	fakeCh := make(chan domain.Chan, len(fake.FakeStreams))

	_, cancel := context.WithCancel(context.Background())
	for i := 0; i < len(fake.FakeStreams); i++ {
		r := repository.NewLoaderRepo(m, &nats.Conn{}, fake.FakeStream)
		s := service.NewLoaderService(&r, fake.FakeStreams[i])

		c := domain.Chan{ChanService: s, Stream: fake.FakeStreams[i]}
		fakeCh <- c
	}

	return fakeCh, cancel
}

func TestOptChan(t *testing.T) {
	ch, cancel := initial(&mongo.Client{})
	r := gin.Default()

	t.Setenv("GIN_MODE", "release")
	h := delivery.NewLoaderHandler(ch, len(fake.FakeStreams), cancel)
	h.Route(r)
	h.Amount = len(fake.FakeStreams)
	h.ChanService = ch

	t.Run("should return true", func(t *testing.T) {
		assert.Equal(t, true, h.OptChan("fakeStream1"))
	})

	t.Run("should return false", func(t *testing.T) {
		assert.Equal(t, false, h.OptChan(fake.FakeStream))
	})
}
