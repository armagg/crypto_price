package health

import (
	"context"
	"crypto_price/pkg/config"
	"crypto_price/pkg/db"
	"crypto_price/pkg/exchanges"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

type HealthCheck struct {
	Name         string        `json:"name"`
	Status       HealthStatus  `json:"status"`
	Message      string        `json:"message,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	Timestamp    time.Time     `json:"timestamp"`
}

type HealthResponse struct {
	Status    HealthStatus  `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Uptime    time.Duration `json:"uptime"`
	Version   string        `json:"version"`
	Checks    []HealthCheck `json:"checks"`
	Services  ServiceHealth `json:"services"`
	System    SystemHealth  `json:"system"`
}

type ServiceHealth struct {
	Redis   HealthCheck `json:"redis"`
	MongoDB HealthCheck `json:"mongodb"`
	Binance HealthCheck `json:"binance"`
	KuCoin  HealthCheck `json:"kucoin"`
}

type SystemHealth struct {
	Memory     MemoryStats `json:"memory"`
	Goroutines int         `json:"goroutines"`
	CGOCalls   int64       `json:"cgo_calls"`
}

type MemoryStats struct {
	Allocated      uint64 `json:"allocated_bytes"`
	TotalAllocated uint64 `json:"total_allocated_bytes"`
	SystemMemory   uint64 `json:"system_bytes"`
	GCCycles       uint32 `json:"gc_cycles"`
	NumGC          uint32 `json:"num_gc"`
}

var startTime = time.Now()

func CheckHealth(ctx context.Context) HealthResponse {
	response := HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime),
		Version:   "1.0.0", 
		Checks:    []HealthCheck{},
	}

	response.Services = checkServices(ctx)
	response.System = checkSystem()

	overallStatus := determineOverallStatus(response.Services)
	response.Status = overallStatus

	return response
}

func checkServices(ctx context.Context) ServiceHealth {
	return ServiceHealth{
		Redis:   checkRedis(ctx),
		MongoDB: checkMongoDB(ctx),
		Binance: checkBinance(ctx),
		KuCoin:  checkKuCoin(ctx),
	}
}

func checkRedis(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:      "redis",
		Timestamp: time.Now(),
	}

	start := time.Now()
	defer func() { check.ResponseTime = time.Since(start) }()

	client, err := db.GetRedisClient()
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("failed to get Redis client: %v", err)
		return check
	}

	_, err = client.Ping(ctx).Result()
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Redis ping failed: %v", err)
		return check
	}

	check.Status = StatusHealthy
	check.Message = "Redis connection successful"
	return check
}

func checkMongoDB(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:      "mongodb",
		Timestamp: time.Now(),
	}

	start := time.Now()
	defer func() { check.ResponseTime = time.Since(start) }()

	client, err := db.GetMongoClient()
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("failed to get MongoDB client: %v", err)
		return check
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("MongoDB ping failed: %v", err)
		return check
	}

	cfg := config.GetConfigs()
	db := client.Database(cfg.MarketDatabase)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	count, err := db.Collection(cfg.ConfigCollection).CountDocuments(ctx, bson.M{})
	if err != nil {
		check.Status = StatusDegraded
		check.Message = fmt.Sprintf("MongoDB query failed: %v", err)
		return check
	}

	check.Status = StatusHealthy
	check.Message = fmt.Sprintf("MongoDB connection successful, found %d documents", count)
	return check
}

func checkBinance(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:      "binance",
		Timestamp: time.Now(),
	}

	start := time.Now()
	defer func() { check.ResponseTime = time.Since(start) }()

	prices, err := exchanges.GetAllBinancePrices()
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Binance API check failed: %v", err)
		return check
	}

	if len(prices) == 0 {
		check.Status = StatusDegraded
		check.Message = "Binance API returned no price data"
		return check
	}

	check.Status = StatusHealthy
	check.Message = fmt.Sprintf("Binance API healthy, returned %d prices", len(prices))
	return check
}

func checkKuCoin(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:      "kucoin",
		Timestamp: time.Now(),
	}

	start := time.Now()
	defer func() { check.ResponseTime = time.Since(start) }()

	symbols := []string{"BTC"}
	prices, err := exchanges.GetPricesKucoin(symbols)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("KuCoin API check failed: %v", err)
		return check
	}

	if len(prices) == 0 {
		check.Status = StatusDegraded
		check.Message = "KuCoin API returned no price data"
		return check
	}

	check.Status = StatusHealthy
	check.Message = fmt.Sprintf("KuCoin API healthy, returned %d prices", len(prices))
	return check
}

func checkSystem() SystemHealth {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemHealth{
		Memory: MemoryStats{
			Allocated:      m.Alloc,
			TotalAllocated: m.TotalAlloc,
			SystemMemory:   m.Sys,
			GCCycles:       m.NumGC,
			NumGC:          m.NumGC,
		},
		Goroutines: runtime.NumGoroutine(),
		CGOCalls:   runtime.NumCgoCall(),
	}
}

func determineOverallStatus(services ServiceHealth) HealthStatus {
	checks := []HealthCheck{
		services.Redis,
		services.MongoDB,
		services.Binance,
		services.KuCoin,
	}

	unhealthyCount := 0
	degradedCount := 0

	for _, check := range checks {
		switch check.Status {
		case StatusUnhealthy:
			unhealthyCount++
		case StatusDegraded:
			degradedCount++
		}
	}

	if unhealthyCount > 0 {
		return StatusUnhealthy
	}
	if degradedCount > 0 {
		return StatusDegraded
	}
	return StatusHealthy
}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	response := CheckHealth(ctx)

	switch response.Status {
	case StatusHealthy:
		w.WriteHeader(http.StatusOK)
	case StatusDegraded:
		w.WriteHeader(http.StatusOK) 
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response)
}

func HandleLiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"uptime":    time.Since(startTime).String(),
	}

	json.NewEncoder(w).Encode(response)
}

func HandleReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check critical services for readiness
	redisCheck := checkRedis(ctx)
	mongoCheck := checkMongoDB(ctx)

	if redisCheck.Status == StatusUnhealthy || mongoCheck.Status == StatusUnhealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		response := map[string]interface{}{
			"status": "not ready",
			"reason": "critical services unhealthy",
			"services": map[string]string{
				"redis":   string(redisCheck.Status),
				"mongodb": string(mongoCheck.Status),
			},
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"status": "ready",
		"services": map[string]string{
			"redis":   string(redisCheck.Status),
			"mongodb": string(mongoCheck.Status),
		},
	}
	json.NewEncoder(w).Encode(response)
}
