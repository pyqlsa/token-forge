// Package cmds provides the implementation backing token-forge's cli.
// This section of the cmds package hold the implementation for the generate
// command.
package cmds

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/pyqlsa/token-forge/internal/bar"
	"github.com/pyqlsa/token-forge/internal/ghtoken"
)

// LocalCmd represents the local collision test cli command.
type LocalCmd struct {
	Globals
	TokenParams
	NumTests uint64 `default:"1" help:"Number of tokens to load into the test token database." short:"t"`
}

type tokenDB map[string]bool

func newtokenDB() tokenDB {
	return make(map[string]bool)
}

// Run the generate tokens command to generate GitHub-like tokens.
func (d *LocalCmd) Run() error {
	if len(d.Prefix) > 0 && !ghtoken.IsValidPrefix(d.Prefix) {
		return fmt.Errorf("prefix '%s' is not a valid token prefix", d.Prefix)
	}

	db := newtokenDB()
	genToken := GenGhTokenFunc(d.Prefix)
	log.Printf("loading %d tokens...", d.NumTests)
	db.populate(d.NumTests, genToken)

	collisions, err := db.testCollisions(generatedTokenSource(d.Prefix, d.NumTokens), d.BatchSize)
	if err != nil {
		return err
	}
	log.Printf("test complete with %d collisions", collisions)

	return nil
}

func (db tokenDB) populate(num uint64, genToken func() *ghtoken.GhToken) {
	for i := uint64(0); i < num; i++ {
		fake := genToken()
		if _, exist := db[fake.FullToken]; !exist {
			// safe to add fresh token
			db[fake.FullToken] = true
		} else {
			log.Printf("observed token collision while generating test set for token: %s", fake.FullToken)
		}
	}
}

func (db tokenDB) testCollisions(source tokenSource, batchSize int) (int32, error) {
	log.Printf("testing w/ %d tokens", source.remaining())
	numCollisions := int32(0)
	tokensLeft := source.remaining()
	progress := bar.NewBar(source.remaining())
	wg := sync.WaitGroup{}
	// kick off the initial batch of workers
	for !source.done() && batchSize > 0 && tokensLeft > 0 {
		wg.Add(1)
		// it's slightly faster to do this in the main thread
		token, err := source.pop()
		if err != nil {
			return numCollisions, fmt.Errorf("error popping token: %w", err)
		}
		go func() {
			if _, exist := db[token.FullToken]; exist {
				atomic.AddInt32(&numCollisions, 1)
				log.Printf("!!! collision: %s", token.FullToken)
			}
			wg.Done()
		}()
		tokensLeft--
		batchSize--
		if err := progress.Inc(); err != nil {
			log.Printf("error adding to the progressbar? %v", err)
		}
	}

	for tokensLeft > 0 {
		// as we're gathering results and the source hasn't been drained, kick off
		// new workers; this makes sure that concurrent workers aren't entirely
		// unbounded
		if !source.done() {
			wg.Add(1)
			// it's slightly faster to do this in the main thread
			token, err := source.pop()
			if err != nil {
				return numCollisions, fmt.Errorf("error popping token: %w", err)
			}
			go func() {
				if _, exist := db[token.FullToken]; exist {
					atomic.AddInt32(&numCollisions, 1)
					log.Printf("!!! collision: %s", token.FullToken)
				}
				wg.Done()
			}()
		}
		tokensLeft--
		if err := progress.Inc(); err != nil {
			log.Printf("error adding to the progressbar? %v", err)
		}
	}
	wg.Wait()
	if err := progress.Finish(); err != nil {
		log.Printf("error finishing the progressbar? %v", err)
	}

	return numCollisions, nil
}
