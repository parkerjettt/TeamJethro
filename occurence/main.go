package main

import (
	"encoding/csv"
	"flag"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alphauslabs/bluectl/pkg/logger"
)

var (
	file = flag.String("file", "", "Input file to process")
	word = flag.String("word", "", "Word to count occurrences")
)

func concurrent() {
	lines, err := os.ReadFile(*file)
	if err != nil {
		logger.Error(err)
		return
	}

	targetWord := *word
	var count int64
	numCpu := runtime.NumCPU()
	wq := make(chan string, numCpu)
	wqdone := make(chan bool)
	var wg sync.WaitGroup

	processor := func(id int) {
		defer func() {
			wg.Done()
		}()
		defer func() {
			wqdone <- true
		}()

		localCount := int64(0)

		reader := csv.NewReader(strings.NewReader(<-wq))
		// Skip the header line
		_, err := reader.Read()
		if err != nil {
			return
		}

		for {
			record, err := reader.Read()
			if err != nil {
				break
			}

			// Iterate over all columns in the record
			for _, field := range record {
				// Trim whitespaces from the field for more accurate matching
				trimmedField := strings.TrimSpace(field)
				if strings.Contains(trimmedField, targetWord) {
					localCount++
					break // Exit the loop if the word is found in any column
				}
			}

			for i := 0; i < 10000; i++ {
				dummy := i % 10
				_ = dummy
			}
		}

		atomic.AddInt64(&count, localCount)
	}

	for i := 0; i < numCpu; i++ {
		wg.Add(1)
		go func(i int) {
			processor(i)
		}(i)
	}

	wq <- string(lines)
	close(wq)

	go func() {
		wg.Wait()
		close(wqdone)
	}()

	// Ensure all goroutines have finished processing before printing the count
	<-wqdone

	logger.Infof("Occurrences of '%s': %v", targetWord, count)
}

func main() {
	flag.Parse()
	if *file == "" || *word == "" {
		logger.Errorf("Both file and word must be specified")
		return
	}

	// Log how long it took
	defer func(begin time.Time) {
		logger.Info("duration:", time.Since(begin))
	}(time.Now())

	concurrent()
}
