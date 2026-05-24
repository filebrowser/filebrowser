package fbhttp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	gopath "path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/filebrowser/filebrowser/v2/settings"
)

const defaultCollaboraTokenTTL = 2 * time.Hour

var collaboraPlaceholder = regexp.MustCompile(`<[^>]+>`)

var collaboraOfficeExtensions = map[string]struct{}{
	"doc": {}, "docx": {}, "docm": {}, "dot": {}, "dotx": {}, "odt": {}, "ott": {}, "rtf": {},
	"xls": {}, "xlsx": {}, "xlsm": {}, "xlt": {}, "xltx": {}, "ods": {}, "ots": {}, "csv": {},
	"ppt": {}, "pptx": {}, "pptm": {}, "pot": {}, "potx": {}, "odp": {}, "otp": {},
	"vsd": {}, "vsdx": {}, "odg": {}, "pdf": {},
}

type collaboraOpenResponse struct {
	URL      string `json:"url"`
	FileID   string `json:"fileID"`
	CanWrite bool   `json:"canWrite"`
	Name     string `json:"name"`
}

type collaboraTestRequest struct {
	Collabora settings.Collabora `json:"collabora"`
}

type collaboraTestResponse struct {
	OK     bool                 `json:"ok"`
	Checks []collaboraTestCheck `json:"checks"`
}

type collaboraTestCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type collaboraDiscovery struct {
	Apps []collaboraDiscoveryApp `xml:"net-zone>app"`
}

type collaboraDiscoveryApp struct {
	Name    string                     `xml:"name,attr"`
	Actions []collaboraDiscoveryAction `xml:"action"`
}

type collaboraDiscoveryAction struct {
	Ext      string `xml:"ext,attr"`
	Name     string `xml:"name,attr"`
	URLSrc   string `xml:"urlsrc,attr"`
	Requires string `xml:"requires,attr"`
}

type wopiTokenClaims struct {
	UserID   uint   `json:"uid"`
	Path     string `json:"path"`
	FileID   string `json:"fid"`
	CanWrite bool   `json:"can_write"`
	jwt.RegisteredClaims
}

var collaboraTestHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	req := &collaboraTestRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return http.StatusBadRequest, err
	}

	cfg := req.Collabora
	cfg.URL = strings.TrimRight(strings.TrimSpace(cfg.URL), "/")
	cfg.PublicURL = strings.TrimRight(strings.TrimSpace(cfg.PublicURL), "/")
	cfg.WOPISecret = strings.TrimSpace(cfg.WOPISecret)
	cfg.TokenTTL = strings.TrimSpace(cfg.TokenTTL)
	if cfg.TokenTTL == "" {
		cfg.TokenTTL = "2h"
	}

	checks := make([]collaboraTestCheck, 0, 8)
	add := func(name, status, message string) {
		checks = append(checks, collaboraTestCheck{Name: name, Status: status, Message: message})
	}

	if !cfg.Enabled {
		add("enabled", "warning", "Collabora integration is disabled. Enable it before testing document editing.")
	} else {
		add("enabled", "success", "Collabora integration is enabled.")
	}

	collaboraURL, err := url.Parse(cfg.URL)
	if cfg.URL == "" || err != nil || collaboraURL.Scheme == "" || collaboraURL.Host == "" {
		add("collabora_url", "error", "Collabora URL must be a valid absolute URL, for example http://192.168.68.135:9980.")
		return renderCollaboraTest(w, r, checks), nil
	}
	if collaboraURL.Scheme != "http" && collaboraURL.Scheme != "https" {
		add("collabora_url", "error", "Collabora URL must use http or https.")
		return renderCollaboraTest(w, r, checks), nil
	}
	add("collabora_url", "success", "Collabora URL format is valid: "+cfg.URL)

	publicURL, err := url.Parse(cfg.PublicURL)
	if cfg.PublicURL == "" || err != nil || publicURL.Scheme == "" || publicURL.Host == "" {
		add("public_url", "error", "Public File Browser URL must be set and must be an absolute URL reachable by the Collabora server.")
		return renderCollaboraTest(w, r, checks), nil
	}
	if publicURL.Scheme != "http" && publicURL.Scheme != "https" {
		add("public_url", "error", "Public File Browser URL must use http or https.")
		return renderCollaboraTest(w, r, checks), nil
	}
	add("public_url", "success", "Public File Browser URL format is valid: "+cfg.PublicURL)

	if strings.TrimSpace(cfg.WOPISecret) == "" {
		add("wopi_secret", "warning", "WOPI token secret is empty. File Browser will fall back to the instance key, but a dedicated long random secret is recommended.")
	} else if len(cfg.WOPISecret) < 32 {
		add("wopi_secret", "warning", "WOPI token secret is set but short. Use a long random value, for example openssl rand -hex 32.")
	} else {
		add("wopi_secret", "success", "WOPI token secret is set.")
	}

	if _, err := time.ParseDuration(cfg.TokenTTL); err != nil {
		add("token_ttl", "error", "WOPI token lifetime is invalid. Use a Go duration such as 2h, 30m, or 1h30m.")
	} else {
		add("token_ttl", "success", "WOPI token lifetime is valid: "+cfg.TokenTTL)
	}

	discovery, err := fetchCollaboraDiscovery(r.Context(), cfg.URL)
	if err != nil {
		add("discovery", "error", err.Error())
		return renderCollaboraTest(w, r, checks), nil
	}
	add("discovery", "success", "Collabora /hosting/discovery is reachable and returned valid XML.")

	required := []string{"docx", "xlsx", "pptx"}
	missing := make([]string, 0)
	for _, ext := range required {
		if !collaboraDiscoveryHasAction(discovery, ext) {
			missing = append(missing, ext)
		}
	}
	if len(missing) > 0 {
		add("office_actions", "error", "Missing Collabora edit/view actions for: "+strings.Join(missing, ", "))
	} else {
		add("office_actions", "success", "Collabora discovery contains actions for docx, xlsx, and pptx.")
	}

	if action, ok := firstCollaboraDiscoveryAction(discovery); ok {
		if actionURL, err := url.Parse(action.URLSrc); err == nil && actionURL.Scheme != "" && actionURL.Host != "" {
			if actionURL.Scheme != collaboraURL.Scheme {
				add("editor_url_scheme", "warning", "Discovery advertises editor URLs with "+actionURL.Scheme+" but the configured Collabora URL uses "+collaboraURL.Scheme+". If the browser shows 'invalid response', fix Collabora ssl.enable/ssl.termination/server_name.")
			} else {
				add("editor_url_scheme", "success", "Discovery editor URL scheme matches the configured Collabora URL.")
			}
		}
	}

	wopiProbeURL := strings.TrimRight(cfg.PublicURL, "/") + "/wopi/files/__collabora_test__?access_token=invalid"
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	probeReq, err := http.NewRequestWithContext(ctx, http.MethodGet, wopiProbeURL, nil)
	if err != nil {
		add("wopi_public_url", "error", "Could not build WOPI probe request: "+err.Error())
	} else {
		probeRes, err := http.DefaultClient.Do(probeReq)
		if err != nil {
			add("wopi_public_url", "warning", "Public File Browser URL was not reachable from this server: "+err.Error()+". Collabora must be able to reach this URL.")
		} else {
			defer probeRes.Body.Close()
			if probeRes.StatusCode == http.StatusUnauthorized || probeRes.StatusCode == http.StatusNotFound {
				add("wopi_public_url", "success", "Public File Browser URL reached the WOPI endpoint. Collabora should use this as aliasgroup1/allowed WOPI host.")
			} else {
				add("wopi_public_url", "warning", fmt.Sprintf("Public File Browser URL responded with HTTP %d during WOPI probe. Collabora may still fail if aliasgroup1/allowed WOPI host is not set to %s.", probeRes.StatusCode, cfg.PublicURL))
			}
		}
	}

	return renderCollaboraTest(w, r, checks), nil
})

func renderCollaboraTest(w http.ResponseWriter, r *http.Request, checks []collaboraTestCheck) int {
	ok := true
	for _, check := range checks {
		if check.Status != "success" {
			ok = false
			break
		}
	}
	_, _ = renderJSON(w, r, collaboraTestResponse{OK: ok, Checks: checks})
	return 0
}

func collaboraDiscoveryHasAction(discovery *collaboraDiscovery, ext string) bool {
	for _, app := range discovery.Apps {
		for _, action := range app.Actions {
			if strings.EqualFold(action.Ext, ext) && action.URLSrc != "" && (strings.EqualFold(action.Name, "edit") || strings.EqualFold(action.Name, "view")) {
				return true
			}
		}
	}
	return false
}

func firstCollaboraDiscoveryAction(discovery *collaboraDiscovery) (collaboraDiscoveryAction, bool) {
	for _, app := range discovery.Apps {
		for _, action := range app.Actions {
			if action.URLSrc != "" {
				return action, true
			}
		}
	}
	return collaboraDiscoveryAction{}, false
}

var collaboraOpenHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	cfg := d.collaboraConfig()
	if !cfg.Enabled || strings.TrimSpace(cfg.URL) == "" {
		return http.StatusNotFound, errors.New("collabora integration is disabled")
	}

	if !d.user.Perm.Download {
		return http.StatusForbidden, nil
	}

	filePath := gopath.Clean("/" + strings.TrimPrefix(r.URL.Query().Get("path"), "/"))
	if filePath == "/" || !d.Check(filePath) {
		return http.StatusForbidden, nil
	}

	info, err := d.user.Fs.Stat(filePath)
	if err != nil {
		return errToStatus(err), err
	}
	if info.IsDir() {
		return http.StatusBadRequest, errors.New("collabora can only open files")
	}

	ext := strings.TrimPrefix(strings.ToLower(gopath.Ext(filePath)), ".")
	if !isCollaboraOfficeExtension(ext) {
		return http.StatusBadRequest, fmt.Errorf("unsupported office extension: %s", ext)
	}

	canWrite := d.user.Perm.Modify
	fileID := wopiFileID(d.user.ID, filePath)
	ttl := collaboraTokenTTL(cfg.TokenTTL)
	token, expiresAt, err := createWOPIToken(d, d.user.ID, filePath, fileID, canWrite, ttl)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	publicURL := collaboraPublicURL(d, r)
	wopiSrc := strings.TrimRight(publicURL, "/") + "/wopi/files/" + url.PathEscape(fileID)

	action, err := collaboraActionForExt(r.Context(), cfg.URL, ext, canWrite)
	if err != nil {
		return http.StatusBadGateway, err
	}

	editorURL := buildCollaboraActionURL(action.URLSrc, wopiSrc, token, expiresAt, canWrite)
	return renderJSON(w, r, collaboraOpenResponse{
		URL:      editorURL,
		FileID:   fileID,
		CanWrite: canWrite,
		Name:     info.Name(),
	})
})

func isCollaboraOfficeExtension(ext string) bool {
	_, ok := collaboraOfficeExtensions[strings.TrimPrefix(strings.ToLower(ext), ".")]
	return ok
}

func collaboraTokenTTL(raw string) time.Duration {
	if strings.TrimSpace(raw) == "" {
		return defaultCollaboraTokenTTL
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return defaultCollaboraTokenTTL
	}
	return d
}

func createWOPIToken(d *data, userID uint, filePath, fileID string, canWrite bool, ttl time.Duration) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)
	claims := wopiTokenClaims{
		UserID:   userID,
		Path:     filePath,
		FileID:   fileID,
		CanWrite: canWrite,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   fmt.Sprintf("%d:%s", userID, filePath),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(wopiSigningKey(d))
	return signed, expiresAt, err
}

func wopiSigningKey(d *data) []byte {
	cfg := d.collaboraConfig()
	if strings.TrimSpace(cfg.WOPISecret) != "" {
		return []byte(cfg.WOPISecret)
	}
	return d.settings.Key
}

func wopiFileID(userID uint, filePath string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d:%s", userID, gopath.Clean("/"+strings.TrimPrefix(filePath, "/")))))
	return hex.EncodeToString(sum[:])
}

func collaboraPublicURL(d *data, r *http.Request) string {
	cfg := d.collaboraConfig()
	if strings.TrimSpace(cfg.PublicURL) != "" {
		return strings.TrimRight(strings.TrimSpace(cfg.PublicURL), "/")
	}

	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}

	return strings.TrimRight(scheme+"://"+host+d.server.BaseURL, "/")
}

func collaboraActionForExt(ctx context.Context, collaboraURL, ext string, canWrite bool) (collaboraDiscoveryAction, error) {
	discovery, err := fetchCollaboraDiscovery(ctx, collaboraURL)
	if err != nil {
		return collaboraDiscoveryAction{}, err
	}

	wanted := []string{"view"}
	if canWrite {
		wanted = []string{"edit", "view"}
	}

	for _, actionName := range wanted {
		for _, app := range discovery.Apps {
			for _, action := range app.Actions {
				if strings.EqualFold(action.Ext, ext) && strings.EqualFold(action.Name, actionName) && action.URLSrc != "" {
					return action, nil
				}
			}
		}
	}

	available := make([]string, 0)
	for _, app := range discovery.Apps {
		for _, action := range app.Actions {
			if action.Ext != "" {
				available = append(available, action.Ext+":"+action.Name)
			}
		}
	}
	sort.Strings(available)
	return collaboraDiscoveryAction{}, fmt.Errorf("no Collabora action found for extension %q; available actions include: %s", ext, strings.Join(available, ", "))
}

func fetchCollaboraDiscovery(ctx context.Context, collaboraURL string) (*collaboraDiscovery, error) {
	base := strings.TrimRight(strings.TrimSpace(collaboraURL), "/")
	if base == "" {
		return nil, errors.New("empty collabora.url")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/hosting/discovery", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Collabora discovery: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("Collabora discovery returned HTTP %d", res.StatusCode)
	}

	var discovery collaboraDiscovery
	if err := xml.NewDecoder(res.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("failed to parse Collabora discovery XML: %w", err)
	}
	return &discovery, nil
}

func buildCollaboraActionURL(urlsrc, wopiSrc, token string, expiresAt time.Time, canWrite bool) string {
	cleaned := collaboraPlaceholder.ReplaceAllString(urlsrc, "")
	separator := "?"
	if strings.Contains(cleaned, "?") {
		if strings.HasSuffix(cleaned, "?") || strings.HasSuffix(cleaned, "&") {
			separator = ""
		} else {
			separator = "&"
		}
	}

	permission := "view"
	if canWrite {
		permission = "edit"
	}

	values := url.Values{}
	values.Set("WOPISrc", wopiSrc)
	values.Set("access_token", token)
	values.Set("access_token_ttl", fmt.Sprintf("%d", expiresAt.UnixMilli()))
	values.Set("permission", permission)

	return cleaned + separator + values.Encode()
}
