package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	clients map[string]*Client
	sync.RWMutex
}

type Client struct {
	requests    []time.Time
	lastUpdated time.Time
}

const (
	AllowedRequests = 10              //cant de requests
	TimeWindow      = 1 * time.Minute // cant de request x minuto (en este caso 10 x 1)
)

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*Client),
	}
}

func (rl *RateLimiter) Limit(c *gin.Context) {
	ip := c.ClientIP()
	client := rl.getClient(ip)
	now := time.Now()

	rl.Lock()
	defer rl.Unlock()

	// Clean up old requests
	client.requests = rl.cleanupOldRequests(client.requests, now)

	// Check if the client has exceeded the allowed number of requests
	if len(client.requests) >= AllowedRequests {
		c.AbortWithStatus(http.StatusTooManyRequests)
		return
	}

	// Add the current request timestamp
	client.requests = append(client.requests, now)
	client.lastUpdated = now

	c.Next()
}

func (rl *RateLimiter) getClient(ip string) *Client {
	rl.RLock()
	client, exists := rl.clients[ip]
	rl.RUnlock()

	if !exists {
		rl.Lock()
		client = &Client{requests: []time.Time{}, lastUpdated: time.Now()}
		rl.clients[ip] = client
		rl.Unlock()
	}
	return client
}

func (rl *RateLimiter) cleanupOldRequests(requests []time.Time, now time.Time) []time.Time {
	var newRequests []time.Time
	for _, requestTime := range requests {
		if now.Sub(requestTime) <= TimeWindow {
			newRequests = append(newRequests, requestTime)
		}
	}
	return newRequests
}
