package audit

import (
	"log"
	"testing"
)

func TestSetupAuditLogger(t *testing.T) {
	auditLogger = nil

	setupAuditLogger()

	if auditLogger != log.Default() {
		t.Error("Audit logger isn't set to the default logger!")
	}
}
