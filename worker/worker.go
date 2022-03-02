package worker

import (
	"bytes"
	"io"
	"net"
	"runtime"
	"sync/atomic"
)

type lock struct {
	i64 int64
}

func (l *lock) Lock() {
	for atomic.CompareAndSwapInt64(&l.i64, 0, 1) {
		runtime.Gosched()
	}
}

func (l *lock) Unlock() {
	atomic.StoreInt64(&l.i64, 0)
}

type WorkerMap struct {
	workers map[string]bool
	count   uint64
	lock    *lock
}

func NewWorkerMap() *WorkerMap {
	return &WorkerMap{
		workers: make(map[string]bool),
		lock:    &lock{},
	}
}

func (w *WorkerMap) Add(url string) {
	w.lock.Lock()
	w.workers[url] = false
	w.count++
	w.lock.Unlock()
}

func (w *WorkerMap) Delete(worker string) {
	w.lock.Lock()
	delete(w.workers, worker)
	w.count--
	w.lock.Unlock()
}

type Result struct {
	Data []byte
	Err  error
}

func (w *WorkerMap) Run(reader io.Reader) <-chan Result {
	ch := make(chan Result)
	go func() {
		for {
			w.lock.Lock()
			if w.count > 0 {
				break
			}
			w.lock.Unlock()
			runtime.Gosched()
		}
		url := ""
		for k, v := range w.workers {
			if !v {
				w.workers[k] = true
				w.count--
				w.lock.Unlock()
				url = k
			}
		}
		w.lock.Unlock()
		defer func() {
			w.lock.Lock()
			w.workers[url] = false
			w.count++
			w.lock.Unlock()
		}()

		client, err := net.Dial("tcp", url)
		if err != nil {
			ch <- Result{Err: err}
			return
		}

		_, err = io.Copy(client, reader)
		if err != nil {
			ch <- Result{Err: err}
			return
		}

		data := bytes.NewBuffer(nil)
		buf := [8192]byte{}
		for {
			n, err := client.Read(buf[:])
			if err != nil {
				if err == io.EOF {
					break
				}
				ch <- Result{Err: err}
				return
			}
			data.Write(buf[:n])
			if n < 8192 {
				break
			}
		}
		ch <- Result{Data: data.Bytes()}
	}()
	return ch
}
