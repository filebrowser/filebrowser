package http

import (
	"fmt"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/tomasen/realip"
)

type LogWriter struct {
	Type         string
	Uri          string
	syslogWriter *syslog.Writer
	fileWriter   *os.File
}

var globalLogWriter *LogWriter = nil

func _syslogConnect(type_ string, host string) *syslog.Writer {
	writer, err := syslog.Dial(type_, host, syslog.LOG_EMERG|syslog.LOG_SYSLOG, "filebrowser")
	if err != nil {
		log.Printf("ERROR: fail to connect to the syslog-%s server: %s", type_, host)
		log.Println(err)
		return nil
	} else {
		return writer
	}
}

func (w *LogWriter) Connect() bool {
	if w.Type == "system" {
		return true
	}
	if w.Type == "file" && w.fileWriter != nil {
		return true
	}
	if (w.Type == "syslog-tcp" || w.Type == "syslog-udp") && w.syslogWriter != nil {
		return true
	}
	if w.Type == "file" {
		w.fileWriter = _openFile(w.Uri)
	} else if w.Type == "syslog-udp" {
		w.syslogWriter = _syslogConnect("udp", w.Uri)
	} else if w.Type == "syslog-tcp" {
		w.syslogWriter = _syslogConnect("tcp", w.Uri)
	}
	return w.syslogWriter != nil || w.fileWriter != nil
}

func _openFile(path string) *os.File {
	d := filepath.Dir(path)
	err := os.MkdirAll(d, 0644)
	if err != nil {
		log.Println("ERROR: fail to create dir " + d)
		return nil
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println("ERROR: fail to open file " + path)
		log.Println(err)
		return nil
	}
	return file
}

func (w *LogWriter) Write(msg string) {
	if w.Type == "system" {
		log.Println(msg)
	} else if w.Type == "file" {
		if w.Connect() {
			w.fileWriter.Write([]byte(msg + "\n"))
		}
	} else if w.Type == "syslog-udp" || w.Type == "syslog-tcp" {
		if w.Connect() {
			w.syslogWriter.Emerg(msg)
		}
	} else {
		log.Println("Unsupported request log output type " + w.Type)
	}
}

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

func _logToOutput(msg string, writer *LogWriter) {
	if writer != nil {
		writer.Write(msg)
	}
}

func _getLogWriter(output string) *LogWriter {
	if output == "system" {
		return &LogWriter{Type: "system"}
	} else if strings.HasPrefix(output, "udp://") {
		return &LogWriter{Type: "syslog-udp", Uri: output[6:]}
	} else if strings.HasPrefix(output, "tcp://") {
		return &LogWriter{Type: "syslog-tcp", Uri: output[6:]}
	} else if strings.HasPrefix(output, "file://") {
		return &LogWriter{Type: "file", Uri: output[7:]}
	}
	log.Fatal("Unsupported log output: " + output)
	return &LogWriter{}
}

func _globalLogWriter(output string) *LogWriter {
	if globalLogWriter == nil {
		globalLogWriter = _getLogWriter(output)
	}
	return globalLogWriter

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
	if log_.status == 0 {
		log_.status = 200
	}
	_globalLogWriter(server.RequestLogOutput).Write(formatLog(server.RequestLogFormat, log_))
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
