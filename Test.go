package main

import (
	"fmt"
	"runtime"
)

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func memConsumed() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats

	fmt.Printf("Alloc = %v KiB", bToKb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v KiB", bToKb(m.TotalAlloc))
	fmt.Printf("\tSys = %v KiB", bToKb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
	return bToKb(m.Sys)
}

func bToKb(b uint64) uint64 {
	return b / 1024
}

func test1() {
	fmt.Println("Started memory")
	fmt.Println("---------------------------------")
	memConsumed()

	fmt.Println("---------------------------------")
	fmt.Println("\t\tModel 1")
	fmt.Println("---------------------------------")
	fmt.Println("Memory before process")
	runtime.GC()
	before1 := memConsumed()
	time1 := Use_model1()
	fmt.Println("Memory After process")
	after1 := memConsumed()

	fmt.Printf("%fkb\n", float64(after1-before1)/1000)
	fmt.Println("Time took ", time1)
}

func test2() {
	fmt.Println("Started memory")
	fmt.Println("---------------------------------")
	memConsumed()

	fmt.Println("---------------------------------")
	fmt.Println("\t\tModel 2")
	fmt.Println("---------------------------------")
	fmt.Println("Memory before process")
	runtime.GC()
	before2 := memConsumed()
	time2 := Use_model2()
	fmt.Println("Memory After process")
	after2 := memConsumed()

	fmt.Printf("%fkb\n", float64(after2-before2)/1000)
	fmt.Println("Time took ", time2)
}

func main() {
	test2()
}
