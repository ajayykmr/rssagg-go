// package main

// import "net/http"

// func handlerReadiness(w http.ResponseWriter, r *http.Request) {
// 	respondWithJSON(w, 200, map[string]string{"status": "ok"})
// }

package main

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := map[string]interface{}{
		"status":     "ok",
		"timestamp":  time.Now().Format(time.RFC3339),
		"service":    "rssagg-go",
		"uptime":     time.Since(startTime).String(),
		"goroutines": runtime.NumGoroutine(),
		"memory": map[string]interface{}{
			"alloc_mb":       float64(memStats.Alloc) / 1024 / 1024,
			"total_alloc_mb": float64(memStats.TotalAlloc) / 1024 / 1024,
			"sys_mb":         float64(memStats.Sys) / 1024 / 1024,
			"num_gc":         memStats.NumGC,
		},
	}

	json.NewEncoder(w).Encode(response)
}
