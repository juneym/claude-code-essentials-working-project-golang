package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsHandler_Returns200(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stats", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStatsHandler_ResponseHasAllFields(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stats", nil)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp StatsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.NotEmpty(t, resp.ServerTimeUTC, "server_time_utc must not be empty")
	assert.Greater(t, resp.SysMem.TotalBytes, uint64(0), "sys_mem.total_bytes must be > 0")
	assert.Greater(t, resp.SysMem.UsedBytes, uint64(0), "sys_mem.used_bytes must be > 0")
	assert.Greater(t, resp.SysMem.UsedPercent, float64(0), "sys_mem.used_percent must be > 0")
	assert.Greater(t, resp.GoMem.HeapAllocBytes, uint64(0), "go_mem.heap_alloc_bytes must be > 0")
	assert.Greater(t, resp.GoMem.HeapInUseBytes, uint64(0), "go_mem.heap_in_use_bytes must be > 0")
	assert.Greater(t, resp.GoMem.TotalAllocBytes, uint64(0), "go_mem.total_alloc_bytes must be > 0")
}

func TestStatsHandler_ServerTimeIsRFC3339UTC(t *testing.T) {
	before := time.Now().UTC().Truncate(time.Second)

	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stats", nil)

	r.ServeHTTP(w, req)

	after := time.Now().UTC().Add(time.Second)

	require.Equal(t, http.StatusOK, w.Code)

	var resp StatsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	parsed, err := time.Parse(time.RFC3339, resp.ServerTimeUTC)
	require.NoError(t, err, "server_time_utc must be valid RFC3339")
	assert.Equal(t, "UTC", parsed.Location().String(), "server_time_utc must be in UTC")
	assert.False(t, parsed.Before(before), "server_time_utc must not be before test start")
	assert.False(t, parsed.After(after), "server_time_utc must not be after test end")
}

func TestStatsHandler_SysMemUsedLessThanTotal(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stats", nil)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp StatsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.LessOrEqual(t, resp.SysMem.UsedBytes, resp.SysMem.TotalBytes,
		"used memory must not exceed total memory")
	assert.LessOrEqual(t, resp.SysMem.UsedPercent, float64(100),
		"used_percent must not exceed 100")
}

func TestStatsHandler_ResponseContentType(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stats", nil)

	r.ServeHTTP(w, req)

	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}
