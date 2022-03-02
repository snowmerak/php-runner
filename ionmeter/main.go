package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"php_runner/worker"
	workerlogger "php_runner/worker/worker_logger"
	"strconv"
	"syscall"

	"github.com/go-www/silverlining"
)

func main() {
	serverPort := flag.Int("port", 80, "Server port")
	numWorker := flag.Int("workers", 1, "Number of worker")
	flag.Parse()

	workerPHP := filepath.Join(".", "worker.php")
	if _, err := os.Stat(workerPHP); os.IsNotExist(err) {
		f, err := os.Create(workerPHP)
		if err != nil {
			panic(err)
		}
		if _, err := f.WriteString(worker.Code); err != nil {
			panic(err)
		}
	}

	workers := worker.NewWorkerMap()

	for i := 0; i < *numWorker; i++ {
		fmt.Println("Start worker", i)
		go func(i int) {
			port := strconv.Itoa(39170 + i)
			cmd := exec.Command("php", "worker.php", port)
			workers.Add("localhost:" + port)
			cmd.Stdin = os.Stdin
			cmd.Stdout = workerlogger.New("[Worker "+strconv.Itoa(i)+"]", os.Stdout)
			cmd.Stderr = workerlogger.New("[Worker "+strconv.Itoa(i)+"]", os.Stderr)
			if err := cmd.Run(); err != nil {
				log.Println(err)
			}
		}(i)
	}

	go func() {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(*serverPort))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Listening on 0.0.0.0:%d\n", *serverPort)

		defer ln.Close()

		srv := silverlining.Server{}

		srv.Handler = func(r *silverlining.Context) {
			reader := r.BodyReader()
			defer reader.Close()

			result := <-workers.Run(reader)

			if result.Err != nil {
				r.WriteFullBody(http.StatusBadRequest, []byte(result.Err.Error()))
				return
			}

			if err := r.WriteFullBody(http.StatusOK, result.Data); err != nil {
				log.Println(err)
			}
		}

		err = srv.Serve(ln)
		if err != nil {
			log.Fatal(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Signal(syscall.SIGTERM))
	<-ch
}
