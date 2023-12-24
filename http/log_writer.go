package http

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	LOG_TYPE_SYSTEM     = "system"
	LOG_TYPE_FILE       = "file"
	LOG_TYPE_SYSLOG_UDP = "syslog-udp"
	LOG_TYPE_SYSLOG_TCP = "syslog-tcp"
)

type LogWriter struct {
	Type         string
	Uri          string
	syslogWriter *SyslogWriter
	fileWriter   *os.File
}

func (w *LogWriter) Connect() bool {
	if w.Type == LOG_TYPE_SYSTEM {
		return true
	}
	if w.Type == LOG_TYPE_FILE && w.fileWriter != nil {
		return true
	}
	if (w.Type == LOG_TYPE_SYSLOG_TCP || w.Type == LOG_TYPE_SYSLOG_UDP) && w.syslogWriter != nil {
		return true
	}
	if w.Type == LOG_TYPE_FILE {
		w.fileWriter = _openFile(w.Uri)
	} else if w.Type == LOG_TYPE_SYSLOG_UDP {
		w.syslogWriter = MakeSyslogWriter("udp", w.Uri)
	} else if w.Type == LOG_TYPE_SYSLOG_TCP {
		w.syslogWriter = MakeSyslogWriter("tcp", w.Uri)
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
	if w.Type == LOG_TYPE_SYSTEM {
		log.Println(msg)
	} else if w.Type == LOG_TYPE_FILE {
		if w.Connect() {
			w.fileWriter.Write([]byte(msg + "\n"))
		}
	} else if w.Type == LOG_TYPE_SYSLOG_TCP || w.Type == LOG_TYPE_SYSLOG_UDP {
		if w.Connect() {
			w.syslogWriter.Write(msg)
		}
	} else {
		log.Println("Unsupported request log output type " + w.Type)
	}
}

func MakeLogWriter(output string) *LogWriter {
	if output == LOG_TYPE_SYSTEM {
		return &LogWriter{Type: LOG_TYPE_SYSTEM}
	} else if strings.HasPrefix(output, "udp://") {
		return &LogWriter{Type: LOG_TYPE_SYSLOG_UDP, Uri: output[6:]}
	} else if strings.HasPrefix(output, "tcp://") {
		return &LogWriter{Type: LOG_TYPE_SYSLOG_TCP, Uri: output[6:]}
	} else if strings.HasPrefix(output, "file://") {
		return &LogWriter{Type: LOG_TYPE_FILE, Uri: output[7:]}
	}
	log.Fatal("Unsupported log output: " + output)
	return &LogWriter{}
}

func (w *LogWriter) Close() {
	if w.fileWriter != nil {
		w.fileWriter.Close()
		w.fileWriter = nil
	}
	if w.syslogWriter != nil {
		w.syslogWriter.Close()
		w.syslogWriter = nil
	}
}
