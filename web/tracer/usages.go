package tracer

import (
	"fmt"
	"runtime"
	"time"
)

func RunMonitoring() {
	for {
		PrintMemUsage()
		time.Sleep(time.Second)
	}
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\t\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\t\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\t\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
