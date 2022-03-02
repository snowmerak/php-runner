package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	workerlogger "php_runner/worker/worker_logger"
	"strconv"
	"syscall"
)

func main() {
	serverPort := flag.Int("port", 80, "Server port")
	numWorker := flag.Int("workers", 1, "Number of worker")
	flag.Parse()

	for i := 0; i < *numWorker; i++ {
		fmt.Println("Start worker", i)
		go func(i int) {
			cmd := exec.Command("php", "worker.php", strconv.Itoa(39170+i))
			cmd.Stdin = os.Stdin
			cmd.Stdout = workerlogger.New("[Worker "+strconv.Itoa(i)+"]", os.Stdout)
			cmd.Stderr = workerlogger.New("[Worker "+strconv.Itoa(i)+"]", os.Stderr)
			if err := cmd.Run(); err != nil {
				panic(err)
			}
		}(i)
	}

	fmt.Printf("Listening on 0.0.0.0:%d\n", *serverPort)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Signal(syscall.SIGTERM))
	<-ch
}
