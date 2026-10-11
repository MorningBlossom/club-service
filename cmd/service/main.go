package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/MorningBlossom/club-service/internal/handler"
	"github.com/MorningBlossom/club-service/internal/infrastructure/redis"
)

func main() {
	log.Println("Starting StudHub Microservice: club-service...")

	// Initialize Redis client and existing club cache.
	redisClient, err := redis.NewClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	err = redisClient.Close()
	if err != nil {
		log.Fatalf("Failed to close Redis connection: %v", err)
	}

	log.Println("Redis connected successfully")

	// Initialize the distributed lock manager using the same
	// underlying Redis client.
	locker, err := redis.NewLocker(redisClient.Client)
	if err != nil {
		log.Fatalf("Failed to initialize Redis locker: %v", err)
	}

	log.Println("Redis locker initialized successfully")

	rateLimiter := redis.NewRateLimiter(redisClient.Client)

	// Health check for the API Gateway.
	http.HandleFunc("/health", handler.HealthCheck)

	// Existing Redis cache test endpoint.
	http.HandleFunc("/redis-test", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := redisClient.SetClub(ctx, "test-club", "Music Club"); err != nil {
			http.Error(
				w,
				"Failed to store value in Redis: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		value, err := redisClient.GetClub(ctx, "test-club")
		if err != nil {
			http.Error(
				w,
				"Failed to get value from Redis: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Redis value: " + value))
	})

	// Test endpoint for the Redis lock.
	http.HandleFunc("/redis-lock-test", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		lock, err := locker.Acquire(ctx, "lock:club-order")
		if err != nil {
			http.Error(
				w,
				"Failed to acquire Redis lock: "+err.Error(),
				http.StatusConflict,
			)
			return
		}

		// Demonstration only: replace this with the protected operation.
		_, _ = w.Write([]byte("Redis lock acquired successfully"))

		releaseCtx, releaseCancel := context.WithTimeout(
			context.Background(),
			2*time.Second,
		)
		defer releaseCancel()

		if err := lock.Release(releaseCtx); err != nil {
			log.Printf("Failed to release Redis lock: %v", err)
		}
	})

	http.HandleFunc("/rate-limit-test", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// Demo only: use the authenticated student ID in production.
		key := "rate-limit:demo-student"

		allowed, err := rateLimiter.Allow(ctx, key, 5, time.Minute)
		if err != nil {
			log.Printf("Rate limiter error: %v", err)
			http.Error(w, "Rate limiter unavailable",
				http.StatusServiceUnavailable)
			return
		}
		if !allowed {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded",
				http.StatusTooManyRequests)
			return
		}

		_, _ = w.Write([]byte("Request allowed"))
	})

	// TODO: Initialize gRPC interface for ML recommendation pipeline.

	port := ":8080"
	log.Printf("Listening on port %s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
