package delivery_test

import (
	"loader/pkg/delivery"
	"loader/pkg/fake"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestGetRestoreInfo(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	t.Setenv("API_TOKEN", "555")
	t.Setenv("GIN_MODE", "release")
	mt.Run("should return false when token is empty", func(mt *mtest.T) {
		ch, cancel := initial(mt.Client)
		r := gin.Default()
		h := delivery.NewLoaderHandler(ch, len(fake.FakeStreams), cancel)
		h.Route(r)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/restore", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, `{"message":"illegal","status":false}`, w.Body.String())
	})

	mt.Run("should return false when token is wrong", func(mt *mtest.T) {
		ch, cancel := initial(mt.Client)

		r := gin.Default()

		h := delivery.NewLoaderHandler(ch, len(fake.FakeStreams), cancel)
		h.Route(r)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/restore", nil)
		req.Header.Set("token", "1234")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, `{"message":"illegal","status":false}`, w.Body.String())
	})

	mt.Run("should return true", func(mt *mtest.T) {

		ch, cancel := initial(mt.Client)
		r := gin.Default()

		c := <-ch
		ch <- c

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "nats-log.restore", mtest.FirstBatch, bson.D{}))

		h := delivery.NewLoaderHandler(ch, len(fake.FakeStreams), cancel)
		h.Route(r)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/restore", nil)
		req.Header.Set("token", "555")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"message":[{"ID":"","CreatedAt":"0001-01-01T00:00:00Z","UpdateAt":"0001-01-01T00:00:00Z","Have":0,"Want":0,"Stream":"","Status":"","RestoreBy":"","Message":""}],"status":true}`, w.Body.String())
	})

}

func TestRestoreBySeq(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	t.Setenv("API_TOKEN", "555")
	t.Setenv("GIN_MODE", "release")
	mt.Run("should return false when token is empty", func(mt *mtest.T) {
		ch, cancel := initial(mt.Client)

		r := gin.Default()

		h := delivery.NewLoaderHandler(ch, len(fake.FakeStreams), cancel)
		h.Route(r)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/restore/sequence", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, `{"message":"Invaild token","status":false}`, w.Body.String())
	})

	mt.Run("should return false when token is wrong", func(mt *mtest.T) {
		ch, cancel := initial(mt.Client)

		r := gin.Default()

		h := delivery.NewLoaderHandler(ch, len(fake.FakeStreams), cancel)
		h.Route(r)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/restore/sequence", nil)
		req.Header.Set("token", "1234")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, `{"message":"Invaild token","status":false}`, w.Body.String())
	})
}

func TestRestoreByTime(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	t.Setenv("API_TOKEN", "555")
	t.Setenv("GIN_MODE", "release")
	mt.Run("should return false when token is empty", func(mt *mtest.T) {
		ch, cancel := initial(mt.Client)

		r := gin.Default()

		h := delivery.NewLoaderHandler(ch, len(fake.FakeStreams), cancel)
		h.Route(r)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/restore/timestamp", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, `{"message":"Invaild token","status":false}`, w.Body.String())
	})

	mt.Run("should return false when token is wrong", func(mt *mtest.T) {
		ch, cancel := initial(mt.Client)

		r := gin.Default()

		h := delivery.NewLoaderHandler(ch, len(fake.FakeStreams), cancel)
		h.Route(r)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/restore/timestamp", nil)
		req.Header.Set("token", "1234")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, `{"message":"Invaild token","status":false}`, w.Body.String())
	})
}
