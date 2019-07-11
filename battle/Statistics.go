package battle

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

// LogEntry represents a battle log entry
type LogEntry struct {
	Time  time.Time
	Event Event
}

// LogWriter allows writing to battle statistics
type LogWriter interface {
	PushEvent(event Event) error
}

// Statistics represents the battle statistics
type Statistics struct {
	lock          *sync.Mutex
	ended         bool
	winnerFaction string
	log           []LogEntry
	logStream     chan LogEntry
}

// NewStatistics creates a new battle statistics instance
func NewStatistics() *Statistics {
	return &Statistics{
		lock:      &sync.Mutex{},
		ended:     false,
		logStream: make(chan LogEntry),
	}
}

// WinnerFaction implements the interface StatisticsReader
func (bstat *Statistics) WinnerFaction() string {
	bstat.lock.Lock()
	winnerFaction := bstat.winnerFaction
	bstat.lock.Unlock()
	return winnerFaction
}

// Log implements the interface StatisticsReader
func (bstat *Statistics) Log() []LogEntry {
	bstat.lock.Lock()
	defer bstat.lock.Unlock()

	log := make([]LogEntry, len(bstat.log))
	copy(log, bstat.log)
	return log
}

// LogStream implements the interface StatisticsReader
func (bstat *Statistics) LogStream() <-chan LogEntry {
	return bstat.logStream
}

// PushEvent pushes a new log entry into the battle statistics
func (bstat *Statistics) PushEvent(event Event) error {
	bstat.lock.Lock()
	defer bstat.lock.Unlock()

	if bstat.ended {
		return errors.New("the battle is already over")
	}

	entry := LogEntry{
		Time:  time.Now(),
		Event: event,
	}

	// Push log entry
	bstat.log = append(bstat.log, entry)

	// Push stream (non-blocking)
	select {
	case bstat.logStream <- entry:
	default:
	}

	return nil
}

// StopRecording stops recording the battle
func (bstat *Statistics) StopRecording() {
	bstat.lock.Lock()
	bstat.ended = true
	bstat.lock.Unlock()
}

// StatisticsReader intefaces battle statistics in read-only mode
type StatisticsReader interface {
	// WinnerFaction returns the name of the winner faction
	WinnerFaction() string

	// Log returns a copy of the battle log
	Log() []LogEntry

	// LogStream returns the log streaming channel
	LogStream() <-chan LogEntry
}
