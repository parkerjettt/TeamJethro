package main

import (
	"flag"
	"runtime"
	"time"
)

var (
	multi = flag.Bool("multi", false, "Set to true to use all CPU/cores")
)

func main() {
	flag.Parse()
	numCpu := 1
	if *multi {
		numCpu = runtime.NumCPU()
	}

	done := make(chan int)
	for i := 0; i < numCpu; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}
			}
		}()
	}

	time.Sleep(time.Second * 10)
	close(done)
}
