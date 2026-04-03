package controller

import "github.com/gin-gonic/gin"

type APIRouter interface {
	Register(g *gin.Engine)
}
