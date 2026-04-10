package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/mem"
)

// GoMemStats contains Go runtime heap statistics.
//
//	@Description	Go runtime memory statistics
type GoMemStats struct {
	HeapAllocBytes  uint64 `json:"heap_alloc_bytes"  example:"2097152"`
	HeapInUseBytes  uint64 `json:"heap_in_use_bytes" example:"3145728"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes" example:"10485760"`
}

// SysMemStats contains OS-level memory statistics.
//
//	@Description	Operating system memory statistics
type SysMemStats struct {
	TotalBytes  uint64  `json:"total_bytes"  example:"17179869184"`
	UsedBytes   uint64  `json:"used_bytes"   example:"8589934592"`
	UsedPercent float64 `json:"used_percent" example:"50.0"`
}

// StatsResponse is the full payload returned by GET /stats.
//
//	@Description	Server and runtime statistics
type StatsResponse struct {
	ServerTimeUTC string      `json:"server_time_utc" example:"2026-04-10T12:00:00Z"`
	SysMem        SysMemStats `json:"sys_mem"`
	GoMem         GoMemStats  `json:"go_mem"`
}

// StatsHandler returns current server time and memory statistics.
//
//	@Summary		Server statistics
//	@Description	Returns UTC server time, OS memory usage (total, in use, in-use %), and Go runtime heap stats
//	@Tags			stats
//	@Produce		json
//	@Success		200	{object}	StatsResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/stats [get]
func StatsHandler(c *gin.Context) {
	// OS-level memory via gopsutil (cross-platform: macOS + Linux)
	vm, err := mem.VirtualMemory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read system memory: " + err.Error()})
		return
	}

	// Go runtime memory — ReadMemStats triggers a brief stop-the-world pause (~10–50µs).
	// Acceptable for an infrequent monitoring endpoint.
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	resp := StatsResponse{
		ServerTimeUTC: time.Now().UTC().Format(time.RFC3339),
		SysMem: SysMemStats{
			TotalBytes:  vm.Total,
			UsedBytes:   vm.Used,
			UsedPercent: vm.UsedPercent,
		},
		GoMem: GoMemStats{
			HeapAllocBytes:  rtm.HeapAlloc,
			HeapInUseBytes:  rtm.HeapInuse,
			TotalAllocBytes: rtm.TotalAlloc,
		},
	}

	c.JSON(http.StatusOK, resp)
}
