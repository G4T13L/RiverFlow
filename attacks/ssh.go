package attacks

import (
	"fmt"
	"net"
	"time"

	"github.com/G4T13L/RiverFlow/base"
	"golang.org/x/crypto/ssh"
)

func SSHconnection(c *base.Credential, params *base.Parameters) (string, error) {
	//SSH basic connection
	var config *ssh.ClientConfig
	if params.TimeLimit == 0 {
		config = &ssh.ClientConfig{
			User: c.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(c.Pass),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else {
		config = &ssh.ClientConfig{
			User: c.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(c.Pass),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         2 * time.Second,
		}
	}

	//Recover function
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
		}
	}()

	t := net.JoinHostPort(params.IpAddress, params.ServerPort)

	_, err := ssh.Dial("tcp", t, config)

	if err != nil {
		return "", err
	}
	return "Connected", err
}
