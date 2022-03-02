package workerlogger

import "io"

type WorkerLogger struct {
	name   string
	writer io.Writer
}

func New(name string, writer io.Writer) io.Writer {
	return &WorkerLogger{
		name:   name,
		writer: writer,
	}
}

func (w *WorkerLogger) Write(p []byte) (int, error) {
	return w.writer.Write(append([]byte(w.name+": "), p...))
}
