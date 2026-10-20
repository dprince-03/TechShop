// Package realtime fans Postgres LISTEN/NOTIFY messages out to in-process subscribers:
// Server-Sent Events streams (order tracking, dispatch, notification bell) and cache
// invalidation (authz, sessions, synonyms). No Redis needed (architecture-decisions #12).
package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Channels the hub listens on.
const (
	ChannelRealtime = "realtime"
	ChannelAuthz    = "authz"
	ChannelSessions = "sessions"
	ChannelSearch   = "search"
)

// Message is a realtime message: Topic selects subscribers (e.g. "order:TS-10482").
type Message struct {
	Topic string          `json:"topic"`
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// Hub listens on Postgres channels and dispatches to subscribers.
type Hub struct {
	pool      *pgxpool.Pool
	logger    *slog.Logger
	mu        sync.RWMutex
	subs      map[string]map[chan Message]struct{}
	listeners map[string][]func(payload string)
}

// NewHub creates a hub.
func NewHub(pool *pgxpool.Pool, logger *slog.Logger) *Hub {
	return &Hub{pool: pool, logger: logger, subs: map[string]map[chan Message]struct{}{}, listeners: map[string][]func(string){}}
}

// OnChannel registers a raw listener for a channel (used for cache invalidation).
func (h *Hub) OnChannel(channel string, f func(payload string)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners[channel] = append(h.listeners[channel], f)
}

// Subscribe returns a channel of messages for a topic and an unsubscribe function.
func (h *Hub) Subscribe(topic string) (<-chan Message, func()) {
	ch := make(chan Message, 16)
	h.mu.Lock()
	if h.subs[topic] == nil {
		h.subs[topic] = map[chan Message]struct{}{}
	}
	h.subs[topic][ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		delete(h.subs[topic], ch)
		h.mu.Unlock()
	}
}

// Run listens until ctx is cancelled, reconnecting on errors.
func (h *Hub) Run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := h.listen(ctx); err != nil && ctx.Err() == nil {
			h.logger.Warn("realtime listener reconnecting", slog.Any("error", err))
			time.Sleep(time.Second)
		}
	}
}

func (h *Hub) listen(ctx context.Context) error {
	conn, err := h.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	for _, c := range []string{ChannelRealtime, ChannelAuthz, ChannelSessions, ChannelSearch} {
		if _, err := conn.Exec(ctx, "listen "+c); err != nil {
			return err
		}
	}
	for {
		n, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}
		h.mu.RLock()
		for _, f := range h.listeners[n.Channel] {
			f(n.Payload)
		}
		h.mu.RUnlock()
		if n.Channel != ChannelRealtime {
			continue
		}
		var m Message
		if json.Unmarshal([]byte(n.Payload), &m) != nil {
			continue
		}
		h.mu.RLock()
		for ch := range h.subs[m.Topic] {
			select {
			case ch <- m:
			default: // slow client: drop; it resumes from the latest snapshot on reconnect
			}
		}
		h.mu.RUnlock()
	}
}

// ServeSSE streams a topic as Server-Sent Events with a 25-second heartbeat. snapshot, if not
// nil, is sent first so a reconnecting client resumes from the current state.
func (h *Hub) ServeSSE(c *gin.Context, topic string, snapshot any) {
	w := c.Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no") // nginx: don't buffer the stream
	w.WriteHeader(http.StatusOK)
	if snapshot != nil {
		b, _ := json.Marshal(snapshot)
		if _, err := fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", b); err != nil {
			return // client went away
		}
	}
	w.Flush()
	msgs, unsubscribe := h.Subscribe(topic)
	defer unsubscribe()
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
			w.Flush()
		case m := <-msgs:
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", m.Event, m.Data); err != nil {
				return
			}
			w.Flush()
		}
	}
}
