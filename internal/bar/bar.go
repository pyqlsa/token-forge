// Package bar provides a progress bar.
package bar

import (
	"fmt"
	"math"
	"os"
	"sync"

	progressbar "github.com/schollz/progressbar/v3"
)

// ProgressBar is a wrapper for schollz/progressbar.
type ProgressBar struct {
	sync.Mutex
	bar   *progressbar.ProgressBar
	buf   int
	count int
}

const defaultBuf = 10000

// common options.
var defaultOpts = []progressbar.Option{
	progressbar.OptionSetWriter(os.Stderr),
	progressbar.OptionOnCompletion(func() {
		fmt.Fprintf(os.Stderr, "\n")
	}),
	progressbar.OptionEnableColorCodes(true),
	progressbar.OptionFullWidth(),
	progressbar.OptionShowCount(),
	progressbar.OptionShowElapsedTimeOnFinish(),
	progressbar.OptionShowIts(),
	progressbar.OptionSetDescription("[cyan]|[reset] Testing... [cyan]|[reset]"),
	progressbar.OptionSetTheme(progressbar.Theme{
		Saucer:        "[cyan]░",
		AltSaucerHead: "[cyan]▒",
		SaucerHead:    "[cyan]▒",
		SaucerPadding: " ",
		BarStart:      "[magenta]|[reset]",
		BarEnd:        "[magenta]|[reset]",
	}),
}

// NewBar returns a new ProgressBar with the maximum set to the given number
// and the default buffer size; returns a spinner if the provided number is
// too big or zero.
func NewBar(num uint64) *ProgressBar {
	return NewBufferedBar(num, defaultBuf)
}

// NewBufferedBar returns a new ProgressBar with the maximum set to the given
// number and the given buffer size; returns a spinner if the provided number is
// too big or zero.
func NewBufferedBar(num uint64, buffer int) *ProgressBar {
	b := buffer

	if b < 1 {
		b = 1
	}

	if num/uint64(b) < 1000 {
		b = 1
	}

	if num == 0 || num > math.MaxInt64 {
		return &ProgressBar{ //nolint:exhaustruct
			bar:   progressbar.NewOptions(-1, defaultOpts...),
			buf:   b,
			count: 0,
		}
	}

	return &ProgressBar{ //nolint:exhaustruct
		bar:   progressbar.NewOptions64(int64(num), defaultOpts...),
		buf:   b,
		count: 0,
	}
}

// Inc increments the progress bar by the amount configured by it's buffer.
func (b *ProgressBar) Inc() error {
	b.Lock()
	defer b.Unlock()
	b.count++

	if b.count >= b.buf {
		b.count = 0

		return b.bar.Add(b.buf) //nolint:wrapcheck
	}

	return nil
}

// Close closes the bar.
func (b *ProgressBar) Close() error {
	b.Lock()
	defer b.Unlock()
	_ = b.bar.Add(b.count)
	b.count = 0

	return b.bar.Close() //nolint:wrapcheck
}

// Finish fills the bar.
func (b *ProgressBar) Finish() error {
	b.Lock()
	defer b.Unlock()
	_ = b.bar.Add(b.count)
	b.count = 0

	return b.bar.Finish() //nolint:wrapcheck
}
