package main

import (
	"fmt"
	"io"
	"sync"
	"time"
)

const maxChatHeartbeats = 40

type chatHeartbeat struct {
	writer   io.Writer
	interval time.Duration

	mu      sync.Mutex
	started bool
	stopped bool
	stop    chan struct{}
	done    chan struct{}
}

func newChatHeartbeat(writer io.Writer, interval time.Duration) *chatHeartbeat {
	return &chatHeartbeat{writer: writer, interval: interval}
}

func (heartbeat *chatHeartbeat) Start() {
	if heartbeat == nil || heartbeat.writer == nil || heartbeat.interval <= 0 {
		return
	}
	heartbeat.mu.Lock()
	defer heartbeat.mu.Unlock()
	if heartbeat.started || heartbeat.stopped {
		return
	}
	heartbeat.started = true
	heartbeat.stop = make(chan struct{})
	heartbeat.done = make(chan struct{})
	started := time.Now()
	go heartbeat.render(started)
}

func (heartbeat *chatHeartbeat) Stop() {
	if heartbeat == nil {
		return
	}
	heartbeat.mu.Lock()
	if !heartbeat.started || heartbeat.stopped {
		heartbeat.mu.Unlock()
		return
	}
	heartbeat.stopped = true
	close(heartbeat.stop)
	done := heartbeat.done
	heartbeat.mu.Unlock()
	<-done
}

func (heartbeat *chatHeartbeat) render(started time.Time) {
	ticker := time.NewTicker(heartbeat.interval)
	defer ticker.Stop()
	defer close(heartbeat.done)
	count := 0
	for {
		select {
		case tick := <-ticker.C:
			if count >= maxChatHeartbeats {
				continue
			}
			elapsed := tick.Sub(started).Milliseconds()
			if elapsed < 0 {
				elapsed = 0
			}
			fmt.Fprintf(heartbeat.writer, "progress\tstate=generating elapsed_ms=%d\n", elapsed)
			count++
		case <-heartbeat.stop:
			return
		}
	}
}
