package audit

import "log"

var auditLogger *log.Logger

func getAuditLogger() *log.Logger {
	if auditLogger == nil {
		auditLogger = log.Default()
	}
	return auditLogger
}
