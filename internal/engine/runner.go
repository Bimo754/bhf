package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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
	return r.RunWithStdin(ctx, scanID, binaryPath, args, nil)
}

// RunWithStdin executes a command with optional stdin streaming.
func (r *Runner) RunWithStdin(ctx context.Context, scanID string, binaryPath string, args []string, stdin io.Reader) ([]string, error) {
	cmd := exec.CommandContext(ctx, binaryPath, args...)

	// Enforce unbuffered stdout/stderr across all Python and C-based subprocesses
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	// Provide non-blocking stdin to prevent interactive CLI tools from blocking
	if stdin != nil {
		cmd.Stdin = stdin
	} else {
		cmd.Stdin = strings.NewReader("")
	}

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
	scanner.Split(scanLinesOrCarriageReturn)

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if strings.TrimSpace(line) == "" {
			continue
		}
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

// scanLinesOrCarriageReturn splits on \n, \r\n, or bare \r to support real-time progress bars.
func scanLinesOrCarriageReturn(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			if i > 0 && data[i-1] == '\r' {
				return i + 1, data[:i-1], nil
			}
			return i + 1, data[:i], nil
		}
		if data[i] == '\r' {
			if i+1 < len(data) {
				if data[i+1] == '\n' {
					return i + 2, data[:i], nil
				}
				return i + 1, data[:i], nil
			}
			if atEOF {
				return i + 1, data[:i], nil
			}
			return 0, nil, nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
