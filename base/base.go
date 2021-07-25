package base

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

var Wg sync.WaitGroup

func Check(e error) {
	if e != nil {
		fmt.Println("[error]", e)
		panic(e)
	}
}

const (
	OK = iota
	CONNECTION_REFUSE
	TIMEOUT
	MAX_FAILED_ATTEMPTS
)

type Credential struct {
	User string
	Pass string
	Pos  string
}

type RequestStructure struct {
	cred    *Credential
	resChan chan<- *Result
}

type Result struct {
	cred      *Credential
	timeTaken time.Duration
	res       string
	err       error
}

type Parameters struct {
	AttackType       string
	IpAddress        string
	ServerPort       string
	TimeLimit        time.Duration
	numberAttempts   int
	UserFile         string
	PassFile         string
	FilterExcludeRes []string
	FilterIncludeRes []string
	FilterExcludeErr []string
	FilterIncludeErr []string
}

func FakeAttack(c *Credential, params Parameters) (string, error) {
	time.Sleep(200 * time.Millisecond)
	return "ok", nil
}

func Worker(
	reqStream <-chan *RequestStructure,
	name string,
	params *Parameters,
	attackFunc func(*Credential, *Parameters) (string, error),
) {
	for {
		if r, ok := <-reqStream; ok {
			t := time.Now()

			// Test credential
			res, err := attackFunc(r.cred, params)
			r.resChan <- &Result{cred: r.cred, timeTaken: time.Since(t), res: res, err: err}
		} else {
			fmt.Println("[worker " + name + "] DEAD")
			break
		}
	}
}

func (c *Credential) manageConcurrently(
	reqStream chan<- *RequestStructure,
	resultStream chan<- *Result,
	params *Parameters,
) {
	responseChan := make(chan *Result)
	var response *Result

	for n := 0; n < params.numberAttempts; n++ {
		// send a request to a worker
		reqStream <- &RequestStructure{cred: c, resChan: responseChan}

		// get raw response
		response = <-responseChan

		//manage the response

		// filters
		if len(params.FilterExcludeRes) > 0 {
			for _, f := range params.FilterExcludeRes {
				if strings.Contains(response.res, f) {
					Wg.Done()
					return
				}
			}
		}
		if len(params.FilterExcludeErr) > 0 {
			for _, f := range params.FilterExcludeErr {
				if response.err != nil && strings.Contains(response.err.Error(), f) {
					Wg.Done()
					return
				}
			}
		}
		if len(params.FilterIncludeRes) > 0 {
			for _, f := range params.FilterIncludeRes {
				if !strings.Contains(response.res, f) {
					Wg.Done()
					return
				}
			}
		}
		if len(params.FilterIncludeErr) > 0 {
			for _, f := range params.FilterIncludeErr {
				if response.err != nil && !strings.Contains(response.err.Error(), f) {
					Wg.Done()
					return
				}
			}
		}

		if response.err == nil {
			break
		}
	}

	//send response result
	resultStream <- response
	Wg.Done()
}

func StartAttacker(
	nWorkers int,
	attackFunc func(*Credential, *Parameters) (string, error),
	params *Parameters,
) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	reqStream := make(chan *RequestStructure)

	for i := 0; i < nWorkers; i++ {
		// fmt.Println("[worker", i, "] Staying Alive")
		go Worker(reqStream, fmt.Sprint(i), params, attackFunc)
	}

	// manage read files types
	credList := Read2Files(params.UserFile, params.PassFile)

	resultStream := make(chan *Result)
	numCredentials := 0

	go func() {
		for c := range credList {
			numCredentials++
			Wg.Add(1)
			go c.manageConcurrently(reqStream, resultStream, params)
		}
		Wg.Wait()
		close(reqStream)
		close(resultStream)
		fmt.Println("[!] Closing the channels")
	}()

	for r := range resultStream {
		if r.err == nil {
			fmt.Printf("[+] %v\t%v\t%v\t%v\n", r.cred.User+":"+r.cred.Pass, r.cred.Pos, r.timeTaken, r.err)
		} else {
			fmt.Printf("[-] %v\t%v\t%v\t%v\n", r.cred.User+":"+r.cred.Pass, r.cred.Pos, r.timeTaken, r.err)
		}
	}
}
