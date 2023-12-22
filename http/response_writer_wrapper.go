package http

import (
	"net/http"
	"time"
)

type _size struct {
	value uint64
}

func (s *_size) get() uint64 {
	return s.value
}
func (s *_size) set(v uint64) {
	s.value = v
}
func (s *_size) add(v uint64) {
	s.value += v
}

type _status struct {
	value int
}

func (s *_status) get() int {
	return s.value
}
func (s *_status) set(v int) {
	s.value = v
}

type ResponseWriterWrapper struct {
	writer http.ResponseWriter
	size   *_size
	status *_status
	time   time.Time
}

func MakeResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		writer: w,
		size:   &_size{value: 0},
		status: &_status{value: 200},
		time:   time.Now(),
	}
}

func (r *ResponseWriterWrapper) Write(data []byte) (int, error) {
	r.size.add(uint64(len(data)))
	return r.writer.Write(data)
}

func (r *ResponseWriterWrapper) Header() http.Header {
	return r.writer.Header()
}

func (r *ResponseWriterWrapper) WriteHeader(statusCode int) {
	r.status.set(statusCode)
	r.writer.WriteHeader(statusCode)
}

func (r *ResponseWriterWrapper) GetSize() uint64 {
	return r.size.get()
}

func (r *ResponseWriterWrapper) GetTime() time.Time {
	return r.time
}

func (r *ResponseWriterWrapper) GetStatus() int {
	return r.status.get()
}

func (r *ResponseWriterWrapper) GetElapsed() float64 {
	return time.Now().Sub(r.time).Seconds()
}
