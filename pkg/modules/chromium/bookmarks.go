package chromium

import (
	"encoding/json"
	"fmt"
)

type legacyBookmark struct {
	Title  string           `json:"title"`
	Page   int              `json:"page"`
	Bold   bool             `json:"bold,omitempty"`
	Italic bool             `json:"italic,omitempty"`
	Color  json.RawMessage  `json:"color,omitempty"`
	Kids   []legacyBookmark `json:"kids,omitempty"`
}

// Keep the pdfcpu payload intact so the Chromium compatibility path preserves
// outline styles and stays independent of the native PDF bookmarks API.
func parseLegacyBookmarks(value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	var tree struct {
		Bookmarks []legacyBookmark `json:"bookmarks"`
	}
	if err := json.Unmarshal([]byte(value), &tree); err != nil {
		return nil, fmt.Errorf("unmarshal bookmarks: %w", err)
	}
	if len(tree.Bookmarks) == 0 {
		return nil, nil
	}
	return []byte(value), nil
}
