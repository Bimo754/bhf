package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
)

// Runner manages low-level subprocess execution with real-time log capturing.
type Runner struct {
	logRepo *database.LogRepo
	hub     *Hub
}

// NewRunner creates a new Runner.
func NewRunner(logRepo *database.LogRepo, hub *Hub) *Runner {
	return &Runner{
		logRepo: logRepo,
		hub:     hub,
	}
}

// Run executes a command, streaming stdout and stderr to the hub and persisting them to the database.
func (r *Runner) Run(ctx context.Context, scanID string, binaryPath string, args []string) ([]string, error) {
	cmd := exec.CommandContext(ctx, binaryPath, args...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	var outputLines []string
	var linesMu sync.Mutex
	var wg sync.WaitGroup

	// Stream stdout reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.streamReader(scanID, models.StreamStdout, stdoutPipe, func(line string) {
			linesMu.Lock()
			outputLines = append(outputLines, line)
			linesMu.Unlock()
		})
	}()

	// Stream stderr reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.streamReader(scanID, models.StreamStderr, stderrPipe, nil)
	}()

	wg.Wait()
	err = cmd.Wait()

	if ctx.Err() == context.Canceled {
		return outputLines, fmt.Errorf("process was manually cancelled")
	}

	if err != nil {
		return outputLines, fmt.Errorf("process exited with error: %w", err)
	}

	return outputLines, nil
}

func (r *Runner) streamReader(scanID string, stream models.LogStream, reader io.Reader, collectFn func(string)) {
	scanner := bufio.NewScanner(reader)
	// Buffer up to 1MB per line for long JSON outputs
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		now := time.Now().UTC()

		if collectFn != nil {
			collectFn(line)
		}

		// Persist line to database
		_ = r.logRepo.Append(&models.LogEntry{
			ScanID:    scanID,
			Stream:    stream,
			Line:      line,
			Timestamp: now,
		})

		// Broadcast to UI subscribers
		r.hub.Broadcast(StreamMessage{
			Type:      "log",
			ScanID:    scanID,
			Stream:    stream,
			Line:      line,
			Timestamp: now,
		})
	}
}
