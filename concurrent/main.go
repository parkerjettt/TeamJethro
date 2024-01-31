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
	cc   = flag.Bool("concurrent", false, "If true, run the concurrent function")
	word = flag.String("word", "", "Word to count occurrences")
)

func sequential() {
	lines, err := os.ReadFile(*file)
	if err != nil {
		logger.Error(err)
		return
	}

	targetWord := *word
	var count int64

	reader := csv.NewReader(strings.NewReader(string(lines)))
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		for _, field := range record {
			if strings.Contains(field, targetWord) {
				atomic.AddInt64(&count, 1)
			}
		}

		for i := 0; i < 10000; i++ {
			dummy := i % 10
			_ = dummy
		}
	}

	logger.Infof("Occurrences of '%s': %v", targetWord, atomic.LoadInt64(&count))
}

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
	wqdone := make(chan bool, numCpu)
	var wg sync.WaitGroup

	processor := func(id int) {
		defer wg.Done()

		localCount := int64(0)

		reader := csv.NewReader(strings.NewReader(<-wq))
		for {
			record, err := reader.Read()
			if err != nil {
				break
			}

			for _, field := range record {
				if strings.Contains(field, targetWord) {
					localCount++
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

	// Create a goroutine to wait for all processors to finish and then close wqdone
	go func() {
		wg.Wait()
		close(wqdone)
	}()

	wq <- string(lines)
	close(wq)

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

	if !*cc {
		sequential()
	} else {
		concurrent()
	}
}
