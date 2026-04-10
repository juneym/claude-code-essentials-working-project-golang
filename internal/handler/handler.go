package handler

import "github.com/gin-gonic/gin"

// ErrorResponse is the envelope for all 4xx/5xx responses.
//
//	@Description	Standard error envelope
type ErrorResponse struct {
	Error string `json:"error" example:"invalid input"`
}

// RegisterRoutes attaches all application routes to the provided engine.
func RegisterRoutes(r *gin.Engine) {
	r.GET("/echo", EchoHandler)
	r.POST("/echo", EchoHandler)
	r.GET("/stats", StatsHandler)
}
