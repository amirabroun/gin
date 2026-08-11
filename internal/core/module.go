package core

import "github.com/gin-gonic/gin"

// Module is the contract every domain module must implement.
// Each module owns its own routes and registers them on the router,
// so the app layer doesn't need to know about module internals.
type Module interface {
	RegisterRoutes(router *gin.Engine)
}
