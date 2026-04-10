package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r)
	return r
}

func TestEchoHandler_GET_ReturnsMessage(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/echo?message=hello+world", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp EchoResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "hello world", resp.Echo)
}

func TestEchoHandler_GET_SpecialCharacters(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/echo?message=hello%21+%F0%9F%91%8B", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp EchoResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "hello! 👋", resp.Echo)
}

func TestEchoHandler_GET_MissingMessage_Returns400(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/echo", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.NotEmpty(t, errResp.Error)
}

func TestEchoHandler_POST_JSONBody(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()

	body, _ := json.Marshal(EchoRequest{Message: "hello from POST"})
	req, _ := http.NewRequest(http.MethodPost, "/echo", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp EchoResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "hello from POST", resp.Echo)
}

func TestEchoHandler_POST_EmptyBody_Returns400(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.NotEmpty(t, errResp.Error)
}

func TestEchoHandler_POST_NoContentType_Returns400(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	// No Content-Type header — ShouldBind falls back to form; no message field → 400
	req, _ := http.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString(`{"message":"test"}`))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEchoHandler_ResponseContentType(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/echo?message=test", nil)

	r.ServeHTTP(w, req)

	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}
