package audit

import "log"

var auditLogger *log.Logger

func init() {
	setupAuditLogger()
}

func setupAuditLogger() {
	auditLogger = log.Default()
}
