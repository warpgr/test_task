package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewHealthCheckController() APIRouter {
	return &healthCheckController{}
}

type healthCheckController struct {
}

func (c *healthCheckController) Register(g *gin.Engine) {
	g.GET("/healthcheck", c.handleHealthCheck)
}

func (c healthCheckController) handleHealthCheck(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}
