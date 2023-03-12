package audit

import "fmt"

// LogResourceActivity prints an entry to the audit log with information about the given event, file and user.
//
// The event should be describing the activity that was performed, e.g. Deletion.
// This results in a log entry like "Deletion of file with path /srv/test.txt by user test (42)".
func LogResourceActivity(info ResourceActivity) {
	message := createLogMessage(info)
	logToAuditLogger(message)
}

func createLogMessage(info ResourceActivity) string {
	return fmt.Sprintf(
		"%v of resource with path %v by user %v (%v)",
		info.Event,
		info.ResourcePath,
		info.User.Username,
		info.User.ID,
	)
}

func logToAuditLogger(message string) {
	getAuditLogger().Println(message)
}
