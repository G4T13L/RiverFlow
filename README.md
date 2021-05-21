# RiverFlow

To run the program use:
```bash
go run model2.go
```

Model program flow

<<<<<<< HEAD
```go
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
```

![model.png](model.png)
=======
![model.png](model.png)
>>>>>>> 2a19d6e59502a9551214c81ca18d673d1144d344
