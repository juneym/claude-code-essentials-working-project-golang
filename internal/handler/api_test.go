// Package handler_test contains black-box API integration tests.
// These tests exercise all endpoints through the full Gin router using
// httptest — no live server or external network calls are required.
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/juneym/hello-world-tasks/internal/handler"
)

// newAPIRouter builds a minimal router identical to the production one
// but without the Swagger UI route (no docs import needed for API tests).
func newAPIRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler.RegisterRoutes(r)
	return r
}

// --- Echo endpoint -----------------------------------------------------------

func TestAPI_Echo_GET_Returns200WithEcho(t *testing.T) {
	r := newAPIRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/echo?message=api+test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "api test", body["echo"])
}

func TestAPI_Echo_POST_Returns200WithEcho(t *testing.T) {
	r := newAPIRouter()
	w := httptest.NewRecorder()

	payload, _ := json.Marshal(map[string]string{"message": "post via api"})
	req, _ := http.NewRequest(http.MethodPost, "/echo", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "post via api", body["echo"])
}

func TestAPI_Echo_GET_MissingMessage_Returns400(t *testing.T) {
	r := newAPIRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/echo", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.NotEmpty(t, body["error"], "error field must be present in 400 response")
}

func TestAPI_Echo_POST_MissingMessage_Returns400(t *testing.T) {
	r := newAPIRouter()
	w := httptest.NewRecorder()

	payload, _ := json.Marshal(map[string]string{})
	req, _ := http.NewRequest(http.MethodPost, "/echo", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.NotEmpty(t, body["error"])
}

func TestAPI_Echo_MessageIsPreservedVerbatim(t *testing.T) {
	const msg = "  spaces  and\ttabs\nand newlines  "
	r := newAPIRouter()
	w := httptest.NewRecorder()

	payload, _ := json.Marshal(map[string]string{"message": msg})
	req, _ := http.NewRequest(http.MethodPost, "/echo", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, msg, body["echo"], "message must be echoed verbatim, including whitespace")
}

// --- Stats endpoint ----------------------------------------------------------

func TestAPI_Stats_Returns200(t *testing.T) {
	r := newAPIRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stats", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestAPI_Stats_ResponseStructure(t *testing.T) {
	r := newAPIRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stats", nil)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Decode into a generic map to verify top-level keys without coupling
	// the test to internal struct types.
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.Contains(t, body, "server_time_utc")
	assert.Contains(t, body, "sys_mem")
	assert.Contains(t, body, "go_mem")

	sysMem, ok := body["sys_mem"].(map[string]any)
	require.True(t, ok, "sys_mem must be an object")
	assert.Contains(t, sysMem, "total_bytes")
	assert.Contains(t, sysMem, "used_bytes")
	assert.Contains(t, sysMem, "used_percent")

	goMem, ok := body["go_mem"].(map[string]any)
	require.True(t, ok, "go_mem must be an object")
	assert.Contains(t, goMem, "heap_alloc_bytes")
	assert.Contains(t, goMem, "heap_in_use_bytes")
	assert.Contains(t, goMem, "total_alloc_bytes")
}

func TestAPI_Stats_ServerTimeIsRecentUTC(t *testing.T) {
	before := time.Now().UTC().Truncate(time.Second)

	r := newAPIRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stats", nil)

	r.ServeHTTP(w, req)

	after := time.Now().UTC().Add(time.Second)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	rawTime, _ := body["server_time_utc"].(string)
	parsed, err := time.Parse(time.RFC3339, rawTime)
	require.NoError(t, err, "server_time_utc must parse as RFC3339")
	assert.False(t, parsed.Before(before), "timestamp must not be before test start")
	assert.False(t, parsed.After(after), "timestamp must not be after test end")
}

func TestAPI_Stats_MemoryValuesArePositive(t *testing.T) {
	r := newAPIRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stats", nil)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	sysMem := body["sys_mem"].(map[string]any)
	assert.Greater(t, sysMem["total_bytes"].(float64), float64(0))
	assert.Greater(t, sysMem["used_bytes"].(float64), float64(0))
	assert.Greater(t, sysMem["used_percent"].(float64), float64(0))

	goMem := body["go_mem"].(map[string]any)
	assert.Greater(t, goMem["heap_alloc_bytes"].(float64), float64(0))
	assert.Greater(t, goMem["heap_in_use_bytes"].(float64), float64(0))
	assert.Greater(t, goMem["total_alloc_bytes"].(float64), float64(0))
}

// --- Not-found behaviour -----------------------------------------------------

func TestAPI_UnknownRoute_Returns404(t *testing.T) {
	r := newAPIRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/does-not-exist", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
