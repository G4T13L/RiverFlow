package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Credential Structures
type login struct {
	user string
	pass string
	pos  string
}

// Dictionary reader, the answer flow to the return channel
func read2files(
	done <-chan interface{},
	userFile string,
	passFile string,
) <-chan interface{} {
	credentialStream := make(chan interface{})
	go func() {
		defer close(credentialStream)

		f1, err := os.Open(userFile)
		check(err)
		f2, err := os.Open(passFile)
		check(err)

		// define closing of files
		defer func() {
			f1.Close()
			f2.Close()
		}()

		// File reader line by line.
		ureader := bufio.NewScanner(f1)
		preader := bufio.NewScanner(f2)

		var userList []string // it is saved in an array so as not to read the user list again
		for i := 0; preader.Scan(); i++ {
			if i == 0 {
				for j := 0; ureader.Scan(); j++ {
					userList = append(userList, ureader.Text())
					select {
					case <-done:
						return
					case credentialStream <- login{user: userList[len(userList)-1],
						pass: preader.Text(),
						pos:  strconv.Itoa(i) + "-" + strconv.Itoa(j)}:
					}
				}
				if err := ureader.Err(); err != nil {
					check(err)
				}
			} else {
				for j, u := range userList {
					select {
					case <-done:
						return
					case credentialStream <- login{user: u,
						pass: preader.Text(),
						pos:  strconv.Itoa(i) + "-" + strconv.Itoa(j)}:
					}
				}
			}
		}
		if err := preader.Err(); err != nil {
			check(err)
		}
	}()
	return credentialStream
}

// Recieve from all channels and return only one channel
func fanIn(
	done <-chan interface{},
	channels ...<-chan string,
) <-chan string {
	var wg sync.WaitGroup
	multiplexedStream := make(chan string, 8)
	multiplex := func(c <-chan string) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case multiplexedStream <- i:
			}
		}
	}
	// Goroutines for readers
	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}
	// Wait for all the reads to complete
	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()
	return multiplexedStream
}

func protocolX( // ssh, ftp, smtp, http, smb
	done <-chan interface{},
	credential login,
	name string,
) string {
	time.Sleep(time.Millisecond * 10) // Replace with a conection for the protocolX
	return fmt.Sprintf("Goroutine %s:\t%s\t%s\t%s", name, credential.user, credential.pass, credential.pos)
}

// Attack job to call protocol X and send the response result
func attack(
	done <-chan interface{},
	readStream <-chan interface{},
	name string,
	protocol func(done <-chan interface{}, credential login, name string) string,
) <-chan string {
	responseStream := make(chan string)
	go func() {
		defer close(responseStream)
		for {
			select {
			case r, ok := <-readStream:
				if !ok {
					responseStream <- fmt.Sprintf("[x] Goroutine %s:\tClose goroutine", name)
					return
				}
				responseStream <- protocol(done, r.(login), name) // send response result of attack
			case <-done:
				return
			}
		}
	}()
	return responseStream
}

func main() {
	done := make(chan interface{})
	defer close(done)

	start := time.Now()
	// create a channel that can receive all credential
	readStream := read2files(done, "users.txt", "passwords.txt")

	//use maximum of green threads
	numWorkers := runtime.NumCPU()

	//fan out
	workers := make([]<-chan string, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i] = attack(done, readStream, strconv.Itoa(i+1), protocolX)
	}

	//fanIn
	for resp := range fanIn(done, workers...) {
		fmt.Println(resp)
	}
	fmt.Printf("Search took: %v", time.Since(start))
}
