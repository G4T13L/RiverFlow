package main

import (
	"context"
	"fmt"
	"time"
)

func attack(i int, retChan chan<- int) {
	time.Sleep(500 * time.Millisecond)
	retChan <- i
	// fmt.Println("send from attack ", i)
}

func generate(ctx context.Context, cancel func()) <-chan int {
	dst := make(chan int)
	n := 1
	go func() {
		defer close(dst)
		for {
			select {
			case <-ctx.Done():
				return
			case dst <- n:
				n++
				// fmt.Println("Send", n)
				if n == 1000 {
					cancel()
				}
			}
		}
	}()
	return dst
}

func worker(ctx context.Context, in <-chan int, timeLimit time.Duration, name int) {
	for job := range in {
		retChan := make(chan int)
		fmt.Println("[worker ", name, "] want to send ", job)
		go attack(job, retChan)
		ctx, _ := context.WithTimeout(ctx, timeLimit)
		select {
		case <-ctx.Done():
			fmt.Println("[worker ", name, "] Time limit for ", job)
			continue
		case res := <-retChan:
			fmt.Println("[worker ", name, "] Result: ", res)
		}

	}
}

func main() {

	ctx, cancel := context.WithCancel(context.TODO())
	getStream := generate(ctx, cancel)
	timeLimit := 250 * time.Millisecond

	for i := 0; i < 5; i++ {
		go worker(context.TODO(), getStream, timeLimit, i)
	}

	time.Sleep(10 * time.Second)
}
