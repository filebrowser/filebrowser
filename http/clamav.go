package fbhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/settings"
)

const (
	clamAVUploadField = "FILES"
	clamAVTimeout     = 10 * time.Minute
)

type clamAVThreatError struct {
	FileName string
	Threat   string
	Details  string
}

func (e *clamAVThreatError) Error() string {
	fileName := strings.TrimSpace(e.FileName)
	if fileName == "" {
		fileName = "uploaded file"
	}

	threat := cleanClamAVMessage(e.Threat, fileName)
	details := cleanClamAVMessage(e.Details, fileName)

	switch {
	case threat != "":
		return fmt.Sprintf("Security scan blocked the upload. Malware was detected in %q. Threat: %s. The file was removed and was not saved.", fileName, threat)
	case details != "":
		return fmt.Sprintf("Security scan blocked the upload. Malware was detected in %q. Scanner details: %s. The file was removed and was not saved.", fileName, details)
	default:
		return fmt.Sprintf("Security scan blocked the upload. Malware was detected in %q. The file was removed and was not saved.", fileName)
	}
}

type clamAVServiceError struct {
	Message string
}

func (e *clamAVServiceError) Error() string {
	return "ClamAV scan failed: " + e.Message
}

func clamAVHTTPStatus(err error) int {
	var threat *clamAVThreatError
	if errors.As(err, &threat) {
		return http.StatusBadRequest
	}

	var svc *clamAVServiceError
	if errors.As(err, &svc) {
		return http.StatusBadGateway
	}

	return errToStatus(err)
}

func scanUploadedFile(ctx context.Context, fs afero.Fs, filePath string, cfg settings.ClamAV) error {
	if !cfg.Enabled {
		return nil
	}

	if strings.TrimSpace(cfg.URL) == "" {
		return &clamAVServiceError{Message: "scanner is enabled but no ClamAV URL is configured"}
	}

	file, err := fs.Open(filePath)
	if err != nil {
		return &clamAVServiceError{Message: fmt.Sprintf("could not open uploaded file for scanning: %v", err)}
	}
	defer file.Close()

	return scanWithClamAV(ctx, cfg, filepath.Base(filePath), file)
}

func testClamAVConnection(ctx context.Context, cfg settings.ClamAV) error {
	if strings.TrimSpace(cfg.URL) == "" {
		return &clamAVServiceError{Message: "no ClamAV URL is configured"}
	}

	return scanWithClamAV(ctx, cfg, "filebrowser-clamav-test.txt", strings.NewReader("filebrowser ClamAV connectivity test\n"))
}

func scanWithClamAV(ctx context.Context, cfg settings.ClamAV, filename string, reader io.Reader) error {
	endpoint := strings.TrimSpace(cfg.URL)
	if endpoint == "" {
		return &clamAVServiceError{Message: "no ClamAV URL is configured"}
	}

	bodyReader, bodyWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(bodyWriter)

	go func() {
		defer bodyWriter.Close()

		part, err := multipartWriter.CreateFormFile(clamAVUploadField, filename)
		if err != nil {
			_ = bodyWriter.CloseWithError(err)
			return
		}

		if _, err = io.Copy(part, reader); err != nil {
			_ = bodyWriter.CloseWithError(err)
			return
		}

		if err = multipartWriter.Close(); err != nil {
			_ = bodyWriter.CloseWithError(err)
			return
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bodyReader)
	if err != nil {
		return &clamAVServiceError{Message: err.Error()}
	}

	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	req.Header.Set("Accept", "application/json, text/plain;q=0.9, */*;q=0.8")
	if cfg.ScanDepth > 0 {
		// Most HTTP ClamAV wrappers ignore per-request recursion settings because
		// ClamAV normally controls archive recursion server-side through MaxRecursion.
		// Sending this as a header keeps compatibility with APIs that reject extra
		// multipart fields while still allowing wrappers that support scan depth to use it.
		req.Header.Set("X-ClamAV-Scan-Depth", strconv.Itoa(cfg.ScanDepth))
	}

	client := &http.Client{Timeout: clamAVTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return &clamAVServiceError{Message: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return &clamAVServiceError{Message: fmt.Sprintf("could not read scanner response: %v", err)}
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if threat := evaluateClamAVResponse(filename, body); threat != nil {
			return threat
		}

		message := strings.TrimSpace(string(body))
		if message == "" {
			message = resp.Status
		} else {
			message = resp.Status + ": " + message
		}
		return &clamAVServiceError{Message: message}
	}

	return evaluateClamAVResponse(filename, body)
}

func evaluateClamAVResponse(filename string, body []byte) error {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil
	}

	var parsed any
	if err := json.Unmarshal(trimmed, &parsed); err == nil {
		if infected, threat, details := jsonContainsInfection(parsed); infected {
			return &clamAVThreatError{FileName: filename, Threat: threat, Details: details}
		}

		if success, ok := jsonBoolField(parsed, "success"); ok && !success {
			message := firstJSONString(parsed, "message", "error", "detail")
			if message == "" {
				message = string(trimmed)
			}
			return &clamAVServiceError{Message: message}
		}

		return nil
	}

	upper := strings.ToUpper(string(trimmed))
	if strings.Contains(upper, "FOUND") || strings.Contains(upper, "INFECTED") {
		return &clamAVThreatError{FileName: filename, Threat: extractClamAVTextThreat(string(trimmed)), Details: string(trimmed)}
	}

	return nil
}

func jsonContainsInfection(value any) (bool, string, string) {
	switch v := value.(type) {
	case map[string]any:
		for key, raw := range v {
			normalized := normalizeClamAVKey(key)
			if isInfectionBoolKey(normalized) {
				if infected, ok := raw.(bool); ok && infected {
					return true, bestJSONThreat(v), firstJSONString(v, "message", "error", "detail", "details", "result", "scan_result")
				}
			}

			if isThreatFieldKey(normalized) {
				if threat := stringFromJSONValue(raw); threat != "" {
					return true, threat, firstJSONString(v, "message", "error", "detail", "details", "result", "scan_result")
				}
			}

			if infected, threat, details := jsonContainsInfection(raw); infected {
				if threat == "" {
					threat = bestJSONThreat(v)
				}
				if details == "" {
					details = firstJSONString(v, "message", "error", "detail", "details", "result", "scan_result")
				}
				return true, threat, details
			}
		}
	case []any:
		for _, item := range v {
			if infected, threat, details := jsonContainsInfection(item); infected {
				return true, threat, details
			}
		}
	}

	return false, "", ""
}

func isInfectionBoolKey(key string) bool {
	switch key {
	case "is_infected", "infected", "virus_found", "malware_found", "malicious", "found":
		return true
	default:
		return false
	}
}

func isThreatFieldKey(key string) bool {
	switch key {
	case "virus", "viruses", "signature", "signatures", "threat", "threats", "malware", "malwares", "virus_name", "threat_name":
		return true
	default:
		return false
	}
}

func normalizeClamAVKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, " ", "_")
	return key
}

func bestJSONThreat(value any) string {
	return firstJSONString(value, "virus", "viruses", "signature", "signatures", "threat", "threats", "malware", "malwares", "virus_name", "threat_name")
}

func stringFromJSONValue(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case []any:
		items := make([]string, 0, len(v))
		for _, item := range v {
			if itemString := stringFromJSONValue(item); itemString != "" {
				items = append(items, itemString)
			}
		}
		return strings.Join(items, ", ")
	case map[string]any:
		return bestJSONThreat(v)
	default:
		return ""
	}
}

func cleanClamAVMessage(message string, fileName string) string {
	message = strings.TrimSpace(message)
	message = strings.Trim(message, "\r\n\t ")
	message = strings.Trim(message, "\"'")
	message = strings.TrimSpace(message)

	if message == "" {
		return ""
	}

	if fileName != "" {
		trimmedFile := strings.TrimSpace(fileName)
		if strings.EqualFold(message, trimmedFile) || strings.EqualFold(message, trimmedFile+":"+trimmedFile) {
			return ""
		}
	}

	return message
}

func extractClamAVTextThreat(message string) string {
	message = strings.TrimSpace(message)
	upper := strings.ToUpper(message)
	idx := strings.LastIndex(upper, " FOUND")
	if idx < 0 {
		return ""
	}

	prefix := strings.TrimSpace(message[:idx])
	if prefix == "" {
		return ""
	}

	if colon := strings.LastIndex(prefix, ":"); colon >= 0 && colon+1 < len(prefix) {
		return strings.TrimSpace(prefix[colon+1:])
	}

	fields := strings.Fields(prefix)
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

func jsonBoolField(value any, field string) (bool, bool) {
	switch v := value.(type) {
	case map[string]any:
		for key, raw := range v {
			if strings.EqualFold(key, field) {
				b, ok := raw.(bool)
				return b, ok
			}
			if b, ok := jsonBoolField(raw, field); ok {
				return b, ok
			}
		}
	case []any:
		for _, item := range v {
			if b, ok := jsonBoolField(item, field); ok {
				return b, ok
			}
		}
	}

	return false, false
}

func firstJSONString(value any, fields ...string) string {
	fieldSet := map[string]struct{}{}
	for _, field := range fields {
		fieldSet[normalizeClamAVKey(field)] = struct{}{}
	}

	switch v := value.(type) {
	case map[string]any:
		for key, raw := range v {
			if _, ok := fieldSet[normalizeClamAVKey(key)]; ok {
				if message := stringFromJSONValue(raw); message != "" {
					return message
				}
			}
		}
		for _, raw := range v {
			if s := firstJSONString(raw, fields...); s != "" {
				return s
			}
		}
	case []any:
		for _, item := range v {
			if s := firstJSONString(item, fields...); s != "" {
				return s
			}
		}
	}

	return ""
}
