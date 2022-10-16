package delivery_test

import (
	"loader/pkg/delivery"
	"loader/pkg/fake"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMetrics(t *testing.T) {
	ch, cancel := initial(&mongo.Client{})
	r := gin.Default()
	h := delivery.NewLoaderHandler(ch, len(fake.FakeStreams), cancel)
	h.Route(r)

	t.Run("", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
