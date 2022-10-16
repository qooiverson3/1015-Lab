package delivery

import (
	"encoding/json"
	"io/ioutil"
	"loader/pkg/domain"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary restore stream by sequence
// @Id 1
// @Tags Follower
// @version 1.0
// @Produce application/json
// @Security ApiKeyAuth
// @Success 202 {object} domain.ResponseOfPostRestoreSuccess
// @failure 403 {object} domain.ResponseOfRestoreForbidden
// @Router /follower [post]
// @Param data body domain.RequestBodyViaSequence true "param"
func (lh *LoaderHandler) Follower(c *gin.Context) {
	filter := domain.Filter{}
	req := domain.RequestBodyViaSequence{}

	if !isTokenExist(c) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "Invaild token",
		})
		log.Printf("[WRN] invaild token")
		return
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invaild body",
		})
		log.Printf("[WRN] invaild request body,%v", err)
		return
	}

	err = json.Unmarshal(jsonData, &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invaild format",
		})
		log.Printf("[WRN] invaild request format,%v", err)
		return
	}

	if !lh.OptChan(req.Stream) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invaild stream",
		})
		return
	}

	if lh.Service.IsRestoreInProgress(req.Stream) {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status":  false,
			"message": "The stream is restoring, try again later",
		})
		return
	}

	if req.StartSeq == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invaild start sequence",
		})
		log.Printf("[WRN] invaild start sequence,%v", err)
		return
	}

	idOrMsg, err := lh.Service.GetTicket(req.Stream)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusAccepted, gin.H{
			"status":  false,
			"message": idOrMsg,
		})
		return
	}

	filter.Seq.Start = req.StartSeq
	filter.Seq.End = req.EndSeq

	c.AbortWithStatusJSON(http.StatusAccepted, gin.H{
		"status": true,
		"ticket": idOrMsg,
	})

	go lh.Service.Restore("seq", idOrMsg, filter, req.Stream, lh.CancelFunc, req.Phase)
}
