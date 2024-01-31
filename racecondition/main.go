package main

import (
	"flag"
	"log"
	"runtime"
	"sync"
)

var (
	lock = flag.Bool("lock", false, "Set to true to add locking")
)

func notSynced() {
	counter := 0
	const num = 5000
	var wg sync.WaitGroup
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			temp := counter
			runtime.Gosched()
			temp++
			counter = temp
			wg.Done()
		}()
	}

	wg.Wait()
	log.Println("count:", counter)
}

func synced() {
	var mtx sync.Mutex
	counter := 0
	const num = 5000
	var wg sync.WaitGroup
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			mtx.Lock()
			temp := counter
			runtime.Gosched()
			temp++
			counter = temp
			mtx.Unlock()
			wg.Done()
		}()
	}

	wg.Wait()
	log.Println("count:", counter)
}

func main() {
	flag.Parse()
	if !*lock {
		notSynced()
	} else {
		synced()
	}
}
