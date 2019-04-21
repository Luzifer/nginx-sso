package splitter

import (
	"bytes"
	"errors"
	"math"
	"sync"
)

var ErrClosedWriter = errors.New("Writing to closed writer prohibited")

// Splitter is a thread-safe writer to split multi-line output at newlines
// and carriage-returns. For example to split program output having progress
// lines in it like ffmpeg does.
type Splitter struct {
	buffer []byte
	lock   sync.Mutex
	output chan []byte

	done bool
}

// New creates a new splitter instance and starts the analyzer routeing inside
func New() *Splitter {
	s := &Splitter{
		buffer: []byte{},
		output: make(chan []byte, 1000),
		done:   false, // Explicit declaration though default
	}

	go s.analyze()

	return s
}

// Close disables the analyzer and prevents further write to the splitter
func (c *Splitter) Close() error {
	c.done = true
	return nil
}

// Subscribe returns a channel containing the output lines
func (c Splitter) Subscribe() <-chan []byte { return c.output }

// Write is a standard implementation of io.Writer returning
// ErrClosedWriter on a write after it got closed
func (c *Splitter) Write(p []byte) (n int, err error) {
	if c.done {
		return 0, ErrClosedWriter
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	c.buffer = append(c.buffer, p...)
	return len(p), nil
}

func (c *Splitter) chunk(i int) {
	c.output <- c.buffer[0:i]
	c.buffer = c.buffer[i+1:]
}

func (c *Splitter) analyze() {
	for !c.done {
		c.lock.Lock()

		splits := []int{}
		for _, chr := range []byte{'\r', '\n'} {
			if i := bytes.IndexByte(c.buffer, chr); i > -1 {
				splits = append(splits, i)
			}
		}

		if i := c.minIntSlice(splits); i > -1 {
			c.chunk(i)
		}

		c.lock.Unlock()
	}

	if len(c.buffer) > 0 {
		c.output <- c.buffer
	}

	close(c.output)
}

func (c *Splitter) minIntSlice(in []int) int {
	if len(in) == 0 {
		return -1
	}

	min := math.MaxInt32
	for _, i := range in {
		if i < min {
			min = i
		}
	}

	return min
}
