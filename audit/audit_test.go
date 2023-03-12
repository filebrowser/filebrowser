package audit

import (
	"log"
	"testing"
)

func TestSetupAuditLogger(t *testing.T) {
	auditLogger = nil

	result := getAuditLogger()

	if result != log.Default() {
		t.Error("Audit logger wasn't initialized with the default logger!")
	}
	if auditLogger != result {
		t.Error("Audit logger wasn't set globally!")
	}
}
