package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/tomasen/realip"
)

type RequestLog struct {
	user          *users.User
	ip            string
	time          time.Time
	request_size  uint64
	response_size uint64
	path          string
	method        string
	status        int
	elapsed       float64
	referer       string
	origin        string
}

func (r *RequestLog) user_name() string {
	if r.user != nil {
		return r.user.Username
	}
	return "-"
}

func (r *RequestLog) user_id() string {
	if r.user != nil {
		return fmt.Sprintf("%d", r.user.ID)
	}
	return "-"
}

func (r *RequestLog) user_scope() string {
	if r.user != nil {
		if r.user.Scope == "" {
			return "."
		}
		return r.user.Scope
	}
	return "-"
}

func (r *RequestLog) time_string() string {
	return r.time.Format(time.RFC3339)
}

// support placeholders:
//
//	%{user_name}
//	%{user_id}
//	%{user_scope}
//	%{ip}
//	%{time}
//	%{request_size}
//	%{response_size}: 0 for Transfer-Encoding=chunked
//	%{path}
//	%{method}
//	%{status}
//	%{elapsed}
//	%{referer}
//	%{origin}
func formatLog(format string, log RequestLog) string {
	format = strings.ReplaceAll(format, "%{user_name}", log.user_name())
	format = strings.ReplaceAll(format, "%{user_id}", log.user_id())
	format = strings.ReplaceAll(format, "%{user_scope}", log.user_scope())
	format = strings.ReplaceAll(format, "%{ip}", log.ip)
	format = strings.ReplaceAll(format, "%{time}", log.time_string())
	format = strings.ReplaceAll(format, "%{request_size}", int2string(log.request_size))
	format = strings.ReplaceAll(format, "%{response_size}", int2string(log.response_size))
	format = strings.ReplaceAll(format, "%{path}", log.path)
	format = strings.ReplaceAll(format, "%{method}", log.method)
	format = strings.ReplaceAll(format, "%{status}", int2string(log.status))
	format = strings.ReplaceAll(format, "%{elapsed}", float2string(log.elapsed))
	format = strings.ReplaceAll(format, "%{referer}", log.referer)
	format = strings.ReplaceAll(format, "%{origin}", log.origin)
	return format
}

func str2uint64(d string) uint64 {
	val, err := strconv.ParseInt(d, 10, 64)
	if err != nil {
		return 0
	}
	return uint64(val)
}

func int2string(val any) string {
	return fmt.Sprintf("%d", val)
}

func float2string(val float64) string {
	return fmt.Sprintf("%f", val)
}

func getRequestSize(r *http.Request) uint64 {
	return uint64(r.ContentLength)
}

type myHandler struct {
	f http.HandlerFunc
}

func (h *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.f(w, r)
}

func _log(writer *ResponseWriterWrapper, r *http.Request, user *users.User, server *settings.Server) {
	log_ := RequestLog{
		user:          user,
		status:        writer.GetStatus(),
		elapsed:       writer.GetElapsed(),
		response_size: writer.GetSize(),
		ip:            realip.FromRequest(r),
		time:          writer.GetTime(),
		request_size:  getRequestSize(r),
		origin:        r.Header.Get("Origin"),
		referer:       r.Header.Get("Referer"),
		path:          r.RequestURI,
		method:        r.Method,
	}
	log.Println(formatLog(server.RequestLogFormat, log_))
}

func RequestLogHandlerFunc(handler http.HandlerFunc, server *settings.Server) http.HandlerFunc {
	if !server.EnableRequestLog {
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := MakeResponseWriterWrapper(w)
		handler(writer, r)
		_log(writer, r, nil, server)
	})
}

func RequestLogHandler(handler http.Handler, server *settings.Server) http.Handler {
	if !server.EnableRequestLog {
		return handler
	}
	return &myHandler{f: func(w http.ResponseWriter, r *http.Request) {
		writer := MakeResponseWriterWrapper(w)
		handler.ServeHTTP(writer, r)
		_log(writer, r, nil, server)
	}}
}

func RequestLogHandleFunc(handle handleFunc, server *settings.Server) handleFunc {
	if !server.EnableRequestLog {
		return handle
	}
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		writer := MakeResponseWriterWrapper(w)
		status, err := handle(writer, r, d)
		writer.status.set(status)
		_log(writer, r, d.user, server)
		return status, err
	}
}
