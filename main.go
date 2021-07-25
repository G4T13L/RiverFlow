package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/G4T13L/RiverFlow/attacks"
	"github.com/G4T13L/RiverFlow/base"
)

func main() {
	numCpus := runtime.NumCPU()
	start := time.Now()
	params := base.Parameters{
		AttackType: "ssh",
		IpAddress:  "192.168.1.42",
		ServerPort: "22",
		UserFile:   "dictionaries/users.txt",
		PassFile:   "dictionaries/passwords.txt",
		// TimeLimit:  2 * time.Second,
	}

	// fmt.Println(params)
	// time.Sleep(5 * time.Second)

	// base.StartAttacker(numCpus, base.FakeAttack, params)
	base.StartAttacker(numCpus, attacks.SSHconnection, &params)
	fmt.Println("Time taken is ", time.Since(start))
}
