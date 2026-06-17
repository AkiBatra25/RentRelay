package storageworker

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// WAL is the Write-Ahead Log — the notebook that survives restarts.
// Every Put and Delete is recorded here as one plain-text line.
// On startup, Replay() reads all lines and rebuilds the in-memory map.
type WAL struct {
	file   *os.File   // the open log file on disk
	writer *bufio.Writer // buffered writer so disk writes are fast
}

// OpenWAL opens (or creates) the log file at the given path.
// If the file already exists, new entries are appended to the end.
func OpenWAL(path string) (*WAL, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open wal %s: %w", path, err)
	}
	return &WAL{file: f, writer: bufio.NewWriter(f)}, nil
}

// LogPut writes one PUT line to the log file.
// Format: PUT <key> <base64(value)> <version>
// We use base64 for value so that any binary data (even with spaces/newlines) is safe.
func (w *WAL) LogPut(key string, value []byte, version int32) error {
	encoded := base64.StdEncoding.EncodeToString(value)
	_, err := fmt.Fprintf(w.writer, "PUT %s %s %d\n", key, encoded, version)
	if err != nil {
		return err
	}
	return w.writer.Flush() // flush immediately so nothing is stuck in buffer
}

// LogDelete writes one DEL line to the log file.
// Format: DEL <key>
func (w *WAL) LogDelete(key string) error {
	_, err := fmt.Fprintf(w.writer, "DEL %s\n", key)
	if err != nil {
		return err
	}
	return w.writer.Flush()
}

// Replay reads the log file from the beginning and returns the final state
// of all key-value pairs, just like if all the Puts and Deletes had happened again.
// This is called once at startup to rebuild the in-memory map.
func (w *WAL) Replay() (map[string]entry, error) {
	// Go back to the start of the file before reading
	if _, err := w.file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("seek wal: %w", err)
	}

	result := make(map[string]entry)
	scanner := bufio.NewScanner(w.file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue // skip blank lines
		}

		parts := strings.Fields(line) // split by whitespace
		if len(parts) < 2 {
			continue // skip malformed lines
		}

		switch parts[0] {
		case "PUT":
			// PUT lines have 4 parts: PUT key base64value version
			if len(parts) < 4 {
				continue
			}
			key := parts[1]
			decoded, err := base64.StdEncoding.DecodeString(parts[2])
			if err != nil {
				continue // skip corrupted lines
			}
			version, err := strconv.ParseInt(parts[3], 10, 32)
			if err != nil {
				continue
			}
			result[key] = entry{value: decoded, version: int32(version)}

		case "DEL":
			// DEL lines have 2 parts: DEL key
			key := parts[1]
			delete(result, key) // remove from the rebuilt map
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read wal: %w", err)
	}

	return result, nil
}

// Close closes the log file.
func (w *WAL) Close() error {
	if err := w.writer.Flush(); err != nil {
		return err
	}
	return w.file.Close()
}
