package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteSofficeProfileConfig(t *testing.T) {
	dir := t.TempDir()
	if err := writeSofficeProfileConfig(dir); err != nil {
		t.Fatalf("writeSofficeProfileConfig: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(dir, "user", "registrymodifications.xcu"))
	if err != nil {
		t.Fatalf("read LibreOffice profile config: %v", err)
	}
	if !strings.Contains(string(body), `oor:name="BlockUntrustedRefererLinks"`) ||
		!strings.Contains(string(body), `<value>true</value>`) {
		t.Fatalf("profile config does not block untrusted linked content: %s", body)
	}
}
