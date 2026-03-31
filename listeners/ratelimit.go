package listeners

import (
	"net"
	"sync"
	"time"
)

const (
	maxConnsPerIPPerMinute = 30
	maxConcurrentConns     = 500
)

var connLimiter = &rateLimiter{
	ipWindows: make(map[string]*ipWindow),
	semaphore: make(chan struct{}, maxConcurrentConns),
}

type rateLimiter struct {
	mu        sync.Mutex
	ipWindows map[string]*ipWindow
	semaphore chan struct{}
}

type ipWindow struct {
	count   int
	resetAt time.Time
}

func (r *rateLimiter) allowConn(remoteAddr string) bool {
	select {
	case r.semaphore <- struct{}{}:
	default:
		return false
	}
	if !r.checkIP(remoteAddr) {
		<-r.semaphore
		return false
	}
	return true
}

func (r *rateLimiter) releaseConn() {
	<-r.semaphore
}

func (r *rateLimiter) allowRequest(remoteAddr string) bool {
	return r.checkIP(remoteAddr)
}

func (r *rateLimiter) checkIP(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	w, ok := r.ipWindows[host]
	if !ok || now.After(w.resetAt) {
		r.ipWindows[host] = &ipWindow{count: 1, resetAt: now.Add(time.Minute)}
		return true
	}
	w.count++
	return w.count <= maxConnsPerIPPerMinute
}
