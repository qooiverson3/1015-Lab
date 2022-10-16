package delivery

import (
	"context"
	"loader/docs"
	"loader/pkg/domain"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type LoaderHandler struct {
	Service     domain.Service
	ChanService chan domain.Chan
	Amount      int
	CancelFunc  context.CancelFunc
}

func NewLoaderHandler(s chan domain.Chan, amount int, cancelFunc context.CancelFunc) *LoaderHandler {
	return &LoaderHandler{
		ChanService: s,
		Amount:      amount,
		CancelFunc:  cancelFunc,
	}
}

func (lh *LoaderHandler) Route(e *gin.Engine) {
	docs.SwaggerInfo.BasePath = "/api/v1"
	if mode := gin.Mode(); mode == gin.DebugMode {
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	e.GET("/metrics", lh.Metrics)

	apiV1 := e.Group("/api/v1")
	apiV1.GET("/restore", lh.GetRestoreInfo)
	apiV1.POST("/restore/sequence", lh.RestoreBySequence)
	apiV1.POST("/restore/timestamp", lh.RestoreByTimestamp)

}

func (lh *LoaderHandler) OptChan(stream string) bool {
	for i := 0; i < lh.Amount; i++ {
		ch := <-lh.ChanService
		lh.ChanService <- ch

		if ch.Stream == stream {
			lh.Service = ch.ChanService
			return true
		}
	}
	return false
}
