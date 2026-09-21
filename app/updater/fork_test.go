//go:build windows || darwin

package updater

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ollama/ollama/app/version"
)

func TestIsNewerVersion(t *testing.T) {
	cases := []struct {
		candidate, current string
		want               bool
	}{
		// The suffix names the series, not the version. Both sides carry it on
		// any release this build would take, and comparing it as text orders
		// "0.34.0-thinkbudget" below "0.34.0".
		{"v0.34.1-thinkbudget", "0.34.0-thinkbudget", true},
		{"v0.34.0-thinkbudget", "0.34.0-thinkbudget", false},
		{"v0.33.9-thinkbudget", "0.34.0-thinkbudget", false},
		{"v0.34.0-thinkbudget", "0.34.0", false},
		{"v0.34.0", "0.34.0-thinkbudget", false},
		{"v0.35.0-thinkbudget", "0.34.9-thinkbudget", true},
		{"v0.34", "0.34.0", false},
		{"v0.34.0.1", "0.34.0", true},
		{"v1.0.0-thinkbudget", "0.99.99-thinkbudget", true},
		// Nothing numeric to compare is not a newer release.
		{"nightly", "0.34.0", false},
		{"", "0.34.0", false},
	}
	for _, tc := range cases {
		if got := isNewerVersion(tc.candidate, tc.current); got != tc.want {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tc.candidate, tc.current, got, tc.want)
		}
	}
}

// A GitHub releases listing carrying one release with the given assets.
func releaseListing(tag string, assets map[string]string) string {
	out := ""
	for name, url := range assets {
		if out != "" {
			out += ","
		}
		out += fmt.Sprintf(`{"name":%q,"browser_download_url":%q}`, name, url)
	}
	return fmt.Sprintf(`[{"tag_name":%q,"draft":false,"prerelease":true,"assets":[%s]}]`, tag, out)
}

func serveListing(t *testing.T, body string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body)) //nolint:errcheck
	}))
	t.Cleanup(server.Close)
	old := UpdateCheckURLBase
	t.Cleanup(func() { UpdateCheckURLBase = old })
	UpdateCheckURLBase = server.URL + "/releases"
}

func TestForkReleaseWithAnInstallerIsAnUpdate(t *testing.T) {
	u := &Updater{}
	serveListing(t, releaseListing("v99.9.9-thinkbudget", map[string]string{
		installerAssetName(): "https://example.invalid/download",
	}))

	available, resp := u.checkForUpdate(t.Context())
	if !available {
		t.Fatal("expected the newer release to be offered")
	}
	if resp.UpdateVersion != "v99.9.9-thinkbudget" {
		t.Errorf("version = %q", resp.UpdateVersion)
	}
	if resp.UpdateURL != "https://example.invalid/download" {
		t.Errorf("url = %q", resp.UpdateURL)
	}
}

func TestForkReleaseWithoutAnInstallerIsNotAnUpdate(t *testing.T) {
	// This is the resting state of the thinking-budget series: the releases
	// carry a binary and the base runtime, and nothing that can install
	// itself. Offering one would stage a download no path could ever apply.
	u := &Updater{}
	serveListing(t, releaseListing("v99.9.9-thinkbudget", map[string]string{
		"ollama-windows-amd64.exe":         "https://example.invalid/bin",
		"ollama-windows-amd64-runtime.zip": "https://example.invalid/rt",
	}))

	if available, _ := u.checkForUpdate(t.Context()); available {
		t.Fatal("a release with no installer is not an update")
	}
}

func TestForkReleaseNotNewerIsNotAnUpdate(t *testing.T) {
	u := &Updater{}
	serveListing(t, releaseListing(version.Version, map[string]string{
		installerAssetName(): "https://example.invalid/download",
	}))

	if available, _ := u.checkForUpdate(t.Context()); available {
		t.Fatal("the running version is not an update to itself")
	}
}

func TestEmptyCheckURLDisablesTheCheck(t *testing.T) {
	// The setting for a machine that must never move on its own. It has to
	// reach no network at all, so there is no server behind this one.
	old := UpdateCheckURLBase
	t.Cleanup(func() { UpdateCheckURLBase = old })
	UpdateCheckURLBase = ""

	u := &Updater{}
	if available, _ := u.checkForUpdate(t.Context()); available {
		t.Fatal("expected no update check to be made")
	}
}

func TestStagedUpdateIsDiscardedWhenAutoUpdateIsOff(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Skip("staging is platform specific")
	}
	ext := ".exe"
	if runtime.GOOS == "darwin" {
		ext = ".zip"
	}
	// Windows' getStagedUpdate also sweeps the pre-transition staging
	// directory under LOCALAPPDATA. Point that somewhere disposable: a test
	// has no business deleting whatever the machine running it had staged.
	t.Setenv("LOCALAPPDATA", t.TempDir())
	UpdateStageDir = t.TempDir()
	bundle := filepath.Join(UpdateStageDir, "etag", "OllamaSetup"+ext)
	if err := os.MkdirAll(filepath.Dir(bundle), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bundle, []byte("installer"), 0o644); err != nil {
		t.Fatal(err)
	}

	old := AutoUpdateAllowed
	t.Cleanup(func() { AutoUpdateAllowed = old })
	AutoUpdateAllowed = func() (bool, error) { return false, nil }

	if IsUpdatePending() {
		t.Error("a declined update is not pending")
	}
	if _, err := os.Stat(bundle); !os.IsNotExist(err) {
		// Left in place it is re-offered at every startup, and applied the
		// first time the setting reads as enabled or fails to read at all.
		t.Errorf("the staged bundle survived: %v", err)
	}
	if err := DoUpgrade(true); err == nil {
		t.Error("expected DoUpgrade to refuse a declined update")
	}
}

func TestUnreadableSettingsRefuseTheUpgradeButKeepTheBundle(t *testing.T) {
	ext := ".exe"
	if runtime.GOOS == "darwin" {
		ext = ".zip"
	}
	// Windows' getStagedUpdate also sweeps the pre-transition staging
	// directory under LOCALAPPDATA. Point that somewhere disposable: a test
	// has no business deleting whatever the machine running it had staged.
	t.Setenv("LOCALAPPDATA", t.TempDir())
	UpdateStageDir = t.TempDir()
	bundle := filepath.Join(UpdateStageDir, "etag", "OllamaSetup"+ext)
	if err := os.MkdirAll(filepath.Dir(bundle), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bundle, []byte("installer"), 0o644); err != nil {
		t.Fatal(err)
	}

	old := AutoUpdateAllowed
	t.Cleanup(func() { AutoUpdateAllowed = old })
	AutoUpdateAllowed = func() (bool, error) { return false, fmt.Errorf("database is locked") }

	if IsUpdatePending() {
		t.Error("a setting that cannot be read is not permission to upgrade")
	}
	if _, err := os.Stat(bundle); err != nil {
		t.Errorf("an unreadable setting is not a decision to discard the download: %v", err)
	}
}

func TestStagedUpdateIsOfferedWhenAutoUpdateIsOn(t *testing.T) {
	ext := ".exe"
	if runtime.GOOS == "darwin" {
		ext = ".zip"
	}
	// Windows' getStagedUpdate also sweeps the pre-transition staging
	// directory under LOCALAPPDATA. Point that somewhere disposable: a test
	// has no business deleting whatever the machine running it had staged.
	t.Setenv("LOCALAPPDATA", t.TempDir())
	UpdateStageDir = t.TempDir()
	bundle := filepath.Join(UpdateStageDir, "etag", "OllamaSetup"+ext)
	if err := os.MkdirAll(filepath.Dir(bundle), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bundle, []byte("installer"), 0o644); err != nil {
		t.Fatal(err)
	}

	old := AutoUpdateAllowed
	t.Cleanup(func() { AutoUpdateAllowed = old })
	AutoUpdateAllowed = func() (bool, error) { return true, nil }

	if !IsUpdatePending() {
		t.Error("expected the staged update to be pending")
	}
	if _, err := os.Stat(bundle); err != nil {
		t.Errorf("the staged bundle was discarded: %v", err)
	}
}
