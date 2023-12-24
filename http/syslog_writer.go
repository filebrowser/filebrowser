package http

import (
	"log"
	"log/syslog"
)

type SyslogWriter struct {
	Type   string
	Host   string
	Writer *syslog.Writer
}

func MakeSyslogWriter(type_ string, host string) *SyslogWriter {
	return &SyslogWriter{
		Type:   type_,
		Host:   host,
		Writer: nil,
	}
}

func (w *SyslogWriter) Write(msg string) {
	if w.Writer == nil {
		w.Writer = syslogConnect(w.Type, w.Host)
	}
	if w.Writer != nil {
		err := w.Writer.Emerg(msg)
		if err != nil {
			log.Println("ERROR: fail to write syslog")
		}
	}
}

func syslogConnect(type_ string, host string) *syslog.Writer {
	writer, err := syslog.Dial(type_, host, syslog.LOG_EMERG|syslog.LOG_KERN, "filebrowser")
	if err != nil {
		log.Printf("ERROR: fail to connect to the syslog-%s server: %s", type_, host)
		log.Println(err)
		return nil
	} else {
		return writer
	}
}

func (w *SyslogWriter) Close() {
	if w.Writer != nil {
		w.Writer.Close()
		w.Writer = nil
	}
}
