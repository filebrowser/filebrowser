package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/users"
)

type RequestLog struct {
	user          *users.User
	ip            string
	time          time.Time
	request_size  int64
	response_size int64
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
		return r.user.Scope
	}
	return "."
}

func (r *RequestLog) time_string() string {
	return r.time.Format(time.RFC3339)
}

func LogRequest(w http.ResponseWriter, r *http.Request, format string, log_ RequestLog) {
	if log_.status == 0 {
		log_.status = 200
	}
	log_.ip = getRealIp(r)
	log_.time = time.Now()
	log_.request_size = r.ContentLength
	if log_.response_size == 0 {
		log_.response_size = parseSize(w.Header().Get("Content-Length"))
	}
	log_.origin = r.Header.Get("Origin")
	log_.referer = r.Header.Get("Referer")
	log_.path = r.URL.Path
	log_.method = r.Method
	log.Println(formatLog(format, log_))
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

func getRealIp(r *http.Request) string {
	remoteAddr := r.RemoteAddr
	forwardedFor := parseFirstItem(r.Header.Get("X-Forwarded-For"))
	realIp := parseFirstItem(r.Header.Get("X-Real-IP"))
	if forwardedFor != "" {
		return forwardedFor
	}
	if realIp != "" {
		return realIp
	}
	return remoteAddr
}

func parseFirstItem(s string) string {
	items := strings.Split(s, ",")
	if len(items) == 0 {
		return ""
	}
	return items[0]
}

func parseSize(d string) int64 {
	val, err := strconv.ParseInt(d, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func int2string(val any) string {
	return fmt.Sprintf("%d", val)
}

func float2string(val float64) string {
	return fmt.Sprintf("%f", val)
}
