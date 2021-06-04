package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

func memConsumed() [3]uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	malloc, totaloc, sysm := bToKb(m.Alloc), bToKb(m.TotalAlloc), bToKb(m.Sys)
	fmt.Printf("Alloc = %v KiB", malloc)
	fmt.Printf("\tTotalAlloc = %v KiB", totaloc)
	fmt.Printf("\tSys = %v KiB", sysm)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
	return [3]uint64{malloc, totaloc, sysm}
}

func bToKb(b uint64) uint64 {
	return b / 1024
}

func test(model int) {
	fmt.Println("Started memory")
	fmt.Println("---------------------------------")
	memConsumed()

	fmt.Println("---------------------------------")
	fmt.Println("\t\tModel " + strconv.Itoa(model))
	fmt.Println("---------------------------------")
	fmt.Println("Memory before process")
	runtime.GC()
	before1 := memConsumed()
	var time1 time.Duration
	if model == 1 {
		time1 = Use_model1()
	} else {
		time1 = Use_model2()
	}
	fmt.Println("Memory After process")
	after1 := memConsumed()

	fmt.Println("---------------------------------")
	fmt.Printf("Memory difference\nAlloc = %v KiB\tTotalAlloc = %v KiB\tSys = %v KiB\n", after1[0]-before1[0], after1[1]-before1[1], after1[2]-before1[2])
	fmt.Println("---------------------------------")
	fmt.Println("[!] Time took ", time1)
}

func main() {
	model, _ := strconv.Atoi(os.Args[1])
	test(model)

}
