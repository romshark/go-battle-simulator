package battle

import (
	"sync"
	"sync/atomic"
	"time"
)

// DynamicTicker implements a thread-safe ticker
// that allows dynamically changing the interval
type DynamicTicker struct {
	c               chan time.Time
	lock            *sync.Mutex
	currentTickerID uint64
	stop            chan struct{}
}

// NewDynamicTicker creates a new dynamic ticker instance
func NewDynamicTicker() *DynamicTicker {
	return &DynamicTicker{
		c:               make(chan time.Time),
		lock:            &sync.Mutex{},
		currentTickerID: 0,
		stop:            nil,
	}
}

// Reset resets the ticker to apply a new interval.
// If the interval is 0 then the time is stopped until it's reset again.
// Reset is thread-safe and can safely be called concurrently
func (tk *DynamicTicker) Reset(newInterval time.Duration) {
	tk.lock.Lock()

	// Stop current ticker
	if tk.stop != nil {
		tk.stop <- struct{}{}
		tk.stop = nil
	}

	if newInterval == 0 {
		// Stop
		tk.lock.Unlock()
		return
	}

	// Start a new ticker
	tickerID := atomic.AddUint64(&tk.currentTickerID, 1)
	ticker := time.NewTicker(newInterval)
	stop := make(chan struct{})
	tk.stop = stop

	tk.lock.Unlock()

	go func() {
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case tm := <-ticker.C:
				if tickerID != atomic.LoadUint64(&tk.currentTickerID) {
					// Avoid firing the tick because this ticker was canceled
					return
				}

				// Fire a tick in a non-blocking way
				select {
				case tk.c <- tm:
				default:
				}
			}
		}
	}()
}

// C returns the ticker channel
func (tk *DynamicTicker) C() <-chan time.Time {
	return tk.c
}
