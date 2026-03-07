package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jojipanackal/poli/internal/model"
)

// slugify converts a name to a filesystem-safe slug.
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	// Remove anything that's not alphanumeric or hyphen
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// PoliHome returns the root poli config directory.
func PoliHome() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".poli")
}

// groupsDir returns the path to the groups directory.
func groupsDir() string {
	return filepath.Join(PoliHome(), "groups")
}

// groupPath returns the path to a specific group directory.
func groupPath(name string) string {
	return filepath.Join(groupsDir(), slugify(name))
}

// requestsDir returns the path to a group's requests directory.
func requestsDir(group string) string {
	return filepath.Join(groupPath(group), "requests")
}

// requestPath returns the path to a specific request file.
func requestPath(group, name string) string {
	return filepath.Join(requestsDir(group), slugify(name)+".json")
}

// --- Group operations ---

// CreateGroup creates a new group directory with metadata.
func CreateGroup(name string) error {
	gp := groupPath(name)

	if _, err := os.Stat(gp); err == nil {
		return fmt.Errorf("group %q already exists", name)
	}

	if err := os.MkdirAll(filepath.Join(gp, "requests"), 0755); err != nil {
		return fmt.Errorf("failed to create group directory: %w", err)
	}

	g := model.Group{
		Name:      name,
		CreatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(gp, "group.json"), data, 0644)
}

// GroupExists checks if a group exists.
func GroupExists(name string) bool {
	_, err := os.Stat(groupPath(name))
	return err == nil
}

// ListGroups returns all group names.
func ListGroups() ([]model.Group, error) {
	dir := groupsDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var groups []model.Group
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		metaPath := filepath.Join(dir, e.Name(), "group.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var g model.Group
		if err := json.Unmarshal(data, &g); err != nil {
			continue
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// --- Request operations ---

// SaveRequest writes a request to disk.
func SaveRequest(group string, req model.Request) error {
	dir := requestsDir(group)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(requestPath(group, req.Name), data, 0644)
}

// LoadRequest reads a request from disk.
func LoadRequest(group, name string) (model.Request, error) {
	var req model.Request

	data, err := os.ReadFile(requestPath(group, name))
	if err != nil {
		return req, fmt.Errorf("request %q not found in group %q", name, group)
	}

	err = json.Unmarshal(data, &req)
	return req, err
}

// ListRequests returns all requests in a group.
func ListRequests(group string) ([]model.Request, error) {
	dir := requestsDir(group)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var reqs []model.Request
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var r model.Request
		if err := json.Unmarshal(data, &r); err != nil {
			continue
		}
		reqs = append(reqs, r)
	}
	return reqs, nil
}

// FindRequest tries to find a request by name (case-insensitive, partial match).
func FindRequest(group, name string) (model.Request, error) {
	// Try exact match first
	req, err := LoadRequest(group, name)
	if err == nil {
		return req, nil
	}

	// Try partial / case-insensitive match
	reqs, err := ListRequests(group)
	if err != nil {
		return model.Request{}, err
	}

	nameLower := strings.ToLower(name)
	for _, r := range reqs {
		if strings.ToLower(r.Name) == nameLower {
			return r, nil
		}
	}
	for _, r := range reqs {
		if strings.Contains(strings.ToLower(r.Name), nameLower) {
			return r, nil
		}
	}

	return model.Request{}, fmt.Errorf("request %q not found in group %q", name, group)
}

// --- Response operations ---

// responsePath returns the path to a request's last response file.
func responsePath(group, name string) string {
	return filepath.Join(requestsDir(group), slugify(name)+".response.json")
}

// SavedResponse holds a persisted HTTP response.
type SavedResponse struct {
	StatusCode int               `json:"status_code"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	DurationMs int64             `json:"duration_ms"`
	Timestamp  time.Time         `json:"timestamp"`
}

// SaveResponse persists the last response for a request.
func SaveResponse(group, requestName string, resp SavedResponse) error {
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(responsePath(group, requestName), data, 0644)
}

// LoadResponse loads the last saved response for a request.
func LoadResponse(group, requestName string) (SavedResponse, error) {
	var resp SavedResponse

	data, err := os.ReadFile(responsePath(group, requestName))
	if err != nil {
		return resp, fmt.Errorf("no saved response for %q — run `poli ping %q` first", requestName, requestName)
	}

	err = json.Unmarshal(data, &resp)
	return resp, err
}
