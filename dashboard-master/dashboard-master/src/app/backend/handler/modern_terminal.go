// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	restful "github.com/emicklei/go-restful/v3"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	EndOfTransmission = "\u0004"
	
	// Timeouts for various operations
	ConnectTimeout    = 30 * time.Second
	HeartbeatInterval = 30 * time.Second
	SessionTimeout    = 10 * time.Minute
)

// ModernTerminalSession implements PtyHandler with proper context management and error handling
type ModernTerminalSession struct {
	id        string
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *slog.Logger
	
	// Communication channels
	bound     chan error
	sockJS    sockjs.Session
	sizeChan  chan remotecommand.TerminalSize
	doneChan  chan struct{}
	
	// State management
	mu        sync.RWMutex
	closed    bool
	lastActivity time.Time
	
	// Error handling
	errorChan chan error
}

// NewModernTerminalSession creates a new terminal session with proper initialization
func NewModernTerminalSession(ctx context.Context, sessionID string, sockJS sockjs.Session, logger *slog.Logger) *ModernTerminalSession {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}

	sessionCtx, cancel := context.WithCancel(ctx)
	
	session := &ModernTerminalSession{
		id:           sessionID,
		ctx:          sessionCtx,
		cancel:       cancel,
		logger:       logger.With("session_id", sessionID),
		bound:        make(chan error, 1),
		sockJS:       sockJS,
		sizeChan:     make(chan remotecommand.TerminalSize, 1),
		doneChan:     make(chan struct{}),
		errorChan:    make(chan error, 1),
		lastActivity: time.Now(),
	}

	// Start heartbeat and session timeout monitoring
	go session.monitor()
	
	return session
}

// Context returns the session's context
func (t *ModernTerminalSession) Context() context.Context {
	return t.ctx
}

// Close safely closes the terminal session
func (t *ModernTerminalSession) Close() {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	t.mu.Unlock()

	t.logger.Info("Closing terminal session")
	t.cancel()
	
	close(t.sizeChan)
	close(t.doneChan)
	close(t.errorChan)
}

// IsClosed returns whether the session is closed
func (t *ModernTerminalSession) IsClosed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.closed
}

// updateActivity updates the last activity timestamp
func (t *ModernTerminalSession) updateActivity() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastActivity = time.Now()
}

// monitor runs session monitoring for heartbeat and timeout
func (t *ModernTerminalSession) monitor() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.mu.RLock()
			lastActivity := t.lastActivity
			t.mu.RUnlock()

			if time.Since(lastActivity) > SessionTimeout {
				t.logger.Warn("Terminal session timed out due to inactivity")
				t.Close()
				return
			}

			// Send heartbeat
			if err := t.sendHeartbeat(); err != nil {
				t.logger.Error("Failed to send heartbeat", "error", err)
				t.Close()
				return
			}
		}
	}
}

// sendHeartbeat sends a heartbeat message to keep the connection alive
func (t *ModernTerminalSession) sendHeartbeat() error {
	if t.IsClosed() {
		return fmt.Errorf("session is closed")
	}

	msg, err := json.Marshal(TerminalMessage{
		Op:   "heartbeat",
		Data: "ping",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat: %w", err)
	}

	return t.sockJS.Send(string(msg))
}

// Next implements TerminalSizeQueue interface with context support
func (t *ModernTerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		t.logger.Debug("Terminal resize", "width", size.Width, "height", size.Height)
		return &size
	case <-t.doneChan:
		t.logger.Debug("Terminal session done")
		return nil
	case <-t.ctx.Done():
		t.logger.Debug("Terminal session context cancelled")
		return nil
	}
}

// Read implements io.Reader interface with proper error handling and context awareness
func (t *ModernTerminalSession) Read(p []byte) (int, error) {
	if t.IsClosed() {
		return copy(p, EndOfTransmission), io.EOF
	}

	// Create a timeout context for the read operation
	readCtx, cancel := context.WithTimeout(t.ctx, 30*time.Second)
	defer cancel()

	// Use a goroutine to make sockJS.Recv() cancellable
	resultChan := make(chan struct {
		data string
		err  error
	}, 1)

	go func() {
		data, err := t.sockJS.Recv()
		select {
		case resultChan <- struct {
			data string
			err  error
		}{data, err}:
		case <-readCtx.Done():
		}
	}()

	select {
	case result := <-resultChan:
		if result.err != nil {
			t.logger.Error("Failed to receive from sockJS", "error", result.err)
			return copy(p, EndOfTransmission), result.err
		}

		var msg TerminalMessage
		if err := json.Unmarshal([]byte(result.data), &msg); err != nil {
			t.logger.Error("Failed to unmarshal terminal message", "error", err, "data", result.data)
			return copy(p, EndOfTransmission), err
		}

		t.updateActivity()

		switch msg.Op {
		case "stdin":
			t.logger.Debug("Received stdin data", "length", len(msg.Data))
			return copy(p, msg.Data), nil
		case "resize":
			t.logger.Debug("Received resize", "rows", msg.Rows, "cols", msg.Cols)
			select {
			case t.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}:
			case <-t.ctx.Done():
				return 0, t.ctx.Err()
			default:
				// Channel full, drop the resize event
				t.logger.Warn("Resize channel full, dropping event")
			}
			return 0, nil
		case "heartbeat":
			// Respond to heartbeat
			t.logger.Debug("Received heartbeat")
			return 0, nil
		default:
			t.logger.Warn("Unknown message type", "op", msg.Op)
			return copy(p, EndOfTransmission), fmt.Errorf("unknown message type '%s'", msg.Op)
		}

	case <-readCtx.Done():
		t.logger.Debug("Read operation timed out")
		return copy(p, EndOfTransmission), readCtx.Err()
	}
}

// Write implements io.Writer interface with context awareness
func (t *ModernTerminalSession) Write(p []byte) (int, error) {
	if t.IsClosed() {
		return 0, fmt.Errorf("session is closed")
	}

	msg, err := json.Marshal(TerminalMessage{
		Op:   "stdout",
		Data: string(p),
	})
	if err != nil {
		t.logger.Error("Failed to marshal stdout message", "error", err)
		return 0, err
	}

	// Create timeout context for send operation
	sendCtx, cancel := context.WithTimeout(t.ctx, 10*time.Second)
	defer cancel()

	// Make sockJS.Send() cancellable
	errChan := make(chan error, 1)
	go func() {
		errChan <- t.sockJS.Send(string(msg))
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.logger.Error("Failed to send stdout message", "error", err)
			return 0, err
		}
		t.updateActivity()
		return len(p), nil
	case <-sendCtx.Done():
		t.logger.Error("Send operation timed out")
		return 0, sendCtx.Err()
	}
}

// Toast sends an out-of-band message to the terminal with context support
func (t *ModernTerminalSession) Toast(message string) error {
	if t.IsClosed() {
		return fmt.Errorf("session is closed")
	}

	msg, err := json.Marshal(TerminalMessage{
		Op:   "toast",
		Data: message,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal toast message: %w", err)
	}

	sendCtx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- t.sockJS.Send(string(msg))
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.logger.Error("Failed to send toast message", "error", err, "message", message)
		}
		return err
	case <-sendCtx.Done():
		return sendCtx.Err()
	}
}

// ModernSessionMap provides thread-safe management of terminal sessions with lifecycle support
type ModernSessionMap struct {
	sessions map[string]*ModernTerminalSession
	mu       sync.RWMutex
	logger   *slog.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	
	// Cleanup management
	cleanupTicker *time.Ticker
	cleanupDone   chan struct{}
}

// NewModernSessionMap creates a new session map with cleanup routines
func NewModernSessionMap(ctx context.Context, logger *slog.Logger) *ModernSessionMap {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}

	childCtx, cancel := context.WithCancel(ctx)
	
	sm := &ModernSessionMap{
		sessions:      make(map[string]*ModernTerminalSession),
		logger:        logger,
		ctx:           childCtx,
		cancel:        cancel,
		cleanupTicker: time.NewTicker(time.Minute),
		cleanupDone:   make(chan struct{}),
	}

	// Start cleanup routine
	go sm.cleanupRoutine()
	
	return sm
}

// Get safely retrieves a session by ID
func (sm *ModernSessionMap) Get(sessionID string) *ModernTerminalSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[sessionID]
}

// Set safely stores a session
func (sm *ModernSessionMap) Set(sessionID string, session *ModernTerminalSession) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[sessionID] = session
	sm.logger.Debug("Added terminal session", "session_id", sessionID, "total_sessions", len(sm.sessions))
}

// Remove safely removes a session
func (sm *ModernSessionMap) Remove(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if session, exists := sm.sessions[sessionID]; exists {
		session.Close()
		delete(sm.sessions, sessionID)
		sm.logger.Debug("Removed terminal session", "session_id", sessionID, "total_sessions", len(sm.sessions))
	}
}

// Close gracefully closes a session with proper status reporting
func (sm *ModernSessionMap) Close(sessionID string, status uint32, reason string) {
	sm.mu.Lock()
	session, exists := sm.sessions[sessionID]
	if exists {
		delete(sm.sessions, sessionID)
	}
	sm.mu.Unlock()

	if !exists {
		sm.logger.Warn("Attempted to close non-existent session", "session_id", sessionID)
		return
	}

	sm.logger.Info("Closing terminal session", 
		"session_id", sessionID, 
		"status", status, 
		"reason", reason)

	// Send close message if session is still active
	if !session.IsClosed() {
		if err := session.sockJS.Close(status, reason); err != nil {
			sm.logger.Error("Failed to close sockJS session", 
				"session_id", sessionID, 
				"error", err)
		}
	}

	session.Close()
}

// Shutdown gracefully shuts down the session map
func (sm *ModernSessionMap) Shutdown() {
	sm.logger.Info("Shutting down terminal session map")
	sm.cancel()

	// Stop cleanup routine
	sm.cleanupTicker.Stop()
	close(sm.cleanupDone)

	// Close all sessions
	sm.mu.Lock()
	sessions := make(map[string]*ModernTerminalSession, len(sm.sessions))
	for k, v := range sm.sessions {
		sessions[k] = v
	}
	sm.mu.Unlock()

	for sessionID, session := range sessions {
		sm.logger.Debug("Shutting down session", "session_id", sessionID)
		session.Close()
	}

	sm.logger.Info("Terminal session map shutdown complete")
}

// cleanupRoutine periodically cleans up inactive sessions
func (sm *ModernSessionMap) cleanupRoutine() {
	defer close(sm.cleanupDone)

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-sm.cleanupTicker.C:
			sm.cleanup()
		}
	}
}

// cleanup removes inactive sessions
func (sm *ModernSessionMap) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	for sessionID, session := range sm.sessions {
		session.mu.RLock()
		lastActivity := session.lastActivity
		closed := session.closed
		session.mu.RUnlock()

		if closed || now.Sub(lastActivity) > SessionTimeout {
			toRemove = append(toRemove, sessionID)
		}
	}

	for _, sessionID := range toRemove {
		if session := sm.sessions[sessionID]; session != nil {
			session.Close()
			delete(sm.sessions, sessionID)
			sm.logger.Info("Cleaned up inactive session", "session_id", sessionID)
		}
	}

	if len(toRemove) > 0 {
		sm.logger.Debug("Cleanup complete", 
			"removed_sessions", len(toRemove), 
			"total_sessions", len(sm.sessions))
	}
}

// GetSessionCount returns the current number of active sessions
func (sm *ModernSessionMap) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}