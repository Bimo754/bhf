package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
)

// ActiveScan represents an ongoing execution session with its cancellation handle.
type ActiveScan struct {
	ScanID string
	Cancel context.CancelFunc
}

// Dispatcher coordinates scan execution and cancellation controls.
type Dispatcher struct {
	mu       sync.RWMutex
	active   map[string]context.CancelFunc
	runner   *Runner
	scanRepo *database.ScanRepo
	hub      *Hub
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(runner *Runner, scanRepo *database.ScanRepo, hub *Hub) *Dispatcher {
	return &Dispatcher{
		active:   make(map[string]context.CancelFunc),
		runner:   runner,
		scanRepo: scanRepo,
		hub:      hub,
	}
}

// RegisterActive registers a cancel function for an active scan session.
func (d *Dispatcher) RegisterActive(scanID string, cancel context.CancelFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.active[scanID] = cancel
}

// Unregister removes an active scan session.
func (d *Dispatcher) Unregister(scanID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.active, scanID)
}

// Cancel stops an actively running scan session.
func (d *Dispatcher) Cancel(scanID string) error {
	d.mu.Lock()
	cancel, exists := d.active[scanID]
	d.mu.Unlock()

	if !exists {
		return fmt.Errorf("scan %s is not actively running", scanID)
	}

	cancel()
	d.Unregister(scanID)

	now := time.Now().UTC()
	_ = d.scanRepo.UpdateStatus(scanID, models.StatusStopped, &now, "Cancelled by user")

	d.hub.Broadcast(StreamMessage{
		Type:      "status",
		ScanID:    scanID,
		Status:    models.StatusStopped,
		Timestamp: now,
	})

	return nil
}

// IsRunning checks whether a scan is currently running.
func (d *Dispatcher) IsRunning(scanID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, exists := d.active[scanID]
	return exists
}
