package base

import (
	"bufio"
	"os"
	"strconv"
)

func Read2Files(
	userFile string,
	passFile string,
) <-chan *Credential {
	credentialStream := make(chan *Credential)
	go func() {
		defer close(credentialStream)

		f1, err := os.Open(userFile)
		Check(err)
		f2, err := os.Open(passFile)
		Check(err)

		// define closing of files
		defer func() {
			f1.Close()
			f2.Close()
		}()

		// File reader line by line.
		uReader := bufio.NewScanner(f1)
		pReader := bufio.NewScanner(f2)
		var userList []string // it is saved in an array so as not to read the user list again
		for i := 0; pReader.Scan(); i++ {
			if i == 0 {
				for j := 0; uReader.Scan(); j++ {
					userList = append(userList, uReader.Text())
					credentialStream <- &Credential{User: userList[len(userList)-1],
						Pass: pReader.Text(),
						Pos:  strconv.Itoa(j) + "-" + strconv.Itoa(i)}
				}
				if err := uReader.Err(); err != nil {
					Check(err)
				}
			} else {
				for j, u := range userList {
					credentialStream <- &Credential{User: u,
						Pass: pReader.Text(),
						Pos:  strconv.Itoa(j) + "-" + strconv.Itoa(i)}
				}
			}
		}
		if err := pReader.Err(); err != nil {
			Check(err)
		}
	}()
	return credentialStream
}