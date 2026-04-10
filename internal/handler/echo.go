package handler

import "net/http"

import "github.com/gin-gonic/gin"

// EchoRequest binds the message from either query string (GET) or JSON body (POST).
//
//	@Description	Echo request payload
type EchoRequest struct {
	Message string `json:"message" form:"message" binding:"required" example:"hello world"`
}

// EchoResponse is the JSON shape returned by the echo endpoint.
//
//	@Description	Echo response envelope
type EchoResponse struct {
	Echo string `json:"echo" example:"hello world"`
}

// EchoHandler echoes the provided message back as JSON.
//
//	@Summary		Echo a message
//	@Description	Accepts a message via query param (GET) or JSON body (POST) and returns it unchanged
//	@Tags			echo
//	@Accept			json
//	@Produce		json
//	@Param			message	query		string		false	"Message to echo (GET)"
//	@Param			body	body		EchoRequest	false	"Message to echo (POST)"
//	@Success		200		{object}	EchoResponse
//	@Failure		400		{object}	ErrorResponse
//	@Router			/echo [get]
//	@Router			/echo [post]
func EchoHandler(c *gin.Context) {
	var req EchoRequest

	// ShouldBind resolves from form/query tags for GET and JSON body for POST.
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, EchoResponse{Echo: req.Message})
}
