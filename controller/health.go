package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (controller *AresController) GetStatus() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	}
}
