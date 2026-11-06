//go:build windows || darwin

package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/ollama/ollama/app/version"
)

// The release asset that carries an installer for this platform.
//
// A release without one is not an update this app can apply, and saying so is
// the point: the thinking-budget builds ship a binary and the base runtime,
// not an installer, so today every check ends here. That is the intended
// resting state rather than a failure -- it is what "do not replace yourself"
// looks like from inside the updater -- and the day a release does carry an
// installer, the check starts working without another change.
func installerAssetName() string {
	switch runtime.GOOS {
	case "windows":
		return "OllamaSetup.exe"
	case "darwin":
		return "Ollama-darwin.zip"
	default:
		return ""
	}
}

// One entry of the GitHub releases listing, cut down to what a check needs.
type forkRelease struct {
	TagName string `json:"tag_name"`
	Draft   bool   `json:"draft"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// Whether `candidate` is a later release than `current`.
//
// Compared on the leading dotted numbers only. The suffix that names the
// series ("-thinkbudget") is deliberately not part of the ordering: it is the
// same on both sides for any release this build would take, and treating it as
// version text would order "0.34.0-thinkbudget" against "0.34.0" by spelling.
//
// Equal numbers are not an update. A missing field on either side compares as
// zero, so 0.34 and 0.34.0 are the same release.
func isNewerVersion(candidate, current string) bool {
	fields := func(v string) []int {
		v = strings.TrimPrefix(strings.TrimSpace(v), "v")
		if i := strings.IndexAny(v, "-+"); i >= 0 {
			v = v[:i]
		}
		var out []int
		for _, f := range strings.Split(v, ".") {
			n, err := strconv.Atoi(f)
			if err != nil {
				return out
			}
			out = append(out, n)
		}
		return out
	}
	a, b := fields(candidate), fields(current)
	for i := 0; i < len(a) || i < len(b); i++ {
		var x, y int
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x != y {
			return x > y
		}
	}
	return false
}

// The newest release in a GitHub releases listing that this build should take.
//
// The listing arrives newest first, so the first entry that is not newer than
// what is running ends the search -- there is nothing older worth taking. A
// draft is not published and is skipped; a pre-release is, and is not, because
// every release in this series is one and filtering them out would filter out
// all of them.
func parseForkListing(body []byte) ([]forkRelease, error) {
	var releases []forkRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func pickForkRelease(body []byte) (bool, UpdateResponse) {
	var none UpdateResponse
	releases, err := parseForkListing(body)
	if err != nil {
		slog.Warn("malformed response checking for update", "error", err)
		return false, none
	}
	asset := installerAssetName()
	if asset == "" {
		slog.Debug("no installer asset is defined for this platform", "os", runtime.GOOS)
		return false, none
	}
	for _, rel := range releases {
		if rel.Draft {
			continue
		}
		if !isNewerVersion(rel.TagName, version.Version) {
			slog.Debug("no update available", "latest", rel.TagName, "running", version.Version)
			return false, none
		}
		for _, a := range rel.Assets {
			if a.Name == asset {
				return true, UpdateResponse{UpdateURL: a.URL, UpdateVersion: rel.TagName}
			}
		}
		// A newer release with nothing this platform can install. Saying so
		// once is worth more than walking back through older ones that, by
		// the ordering above, this build would not take either.
		slog.Info("newer release carries no installer for this platform", "release", rel.TagName, "want", asset)
		return false, none
	}
	return false, none
}

// Whether a newer release of this build's own series is available.
//
// Two answer shapes are accepted, because two kinds of endpoint are worth
// pointing this at. A JSON array is a GitHub releases listing -- what the
// default URL returns, and what the fork publishes. A JSON object is the
// single `{"version": ..., "url": ...}` an update service returns, which is
// what upstream's own endpoint speaks and what a private mirror is easiest to
// build; it is taken at its word, since an endpoint named explicitly has
// already decided what this machine should run.
func checkForUpdateFromFork(ctx context.Context) (bool, UpdateResponse) {
	var none UpdateResponse
	if UpdateCheckURLBase == "" {
		slog.Debug("update checks are disabled for this build")
		return false, none
	}

	req, err := newForkRequest(ctx)
	if err != nil {
		slog.Warn("failed to check for update", "error", err)
		return false, none
	}
	slog.Debug("checking for available update", "requestURL", UpdateCheckURLBase)
	body, err := fetchForkListing(req)
	if err != nil {
		slog.Warn("failed to check for update", "error", err)
		return false, none
	}
	if body == nil {
		// 204: nothing to offer, and nothing to parse.
		return false, none
	}

	var found bool
	var updateResp UpdateResponse
	if len(strings.TrimLeft(string(body), " \t\r\n")) > 0 && strings.TrimLeft(string(body), " \t\r\n")[0] == '[' {
		found, updateResp = pickForkRelease(body)
	} else {
		if err := json.Unmarshal(body, &updateResp); err != nil {
			slog.Warn("malformed response checking for update", "error", err)
			return false, none
		}
		found = updateResp.UpdateURL != ""
	}
	if !found {
		return false, none
	}
	slog.Info("New update available at " + updateResp.UpdateURL)
	return true, updateResp
}

// The request this build makes to its update endpoint.
func newForkRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, UpdateCheckURLBase, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", fmt.Sprintf("ollama/%s %s Go/%s %s", version.Version, runtime.GOARCH, runtime.Version(), UserAgentOS))
	return req, nil
}

// The body of the update endpoint's answer, or nil when it has nothing to say.
func fetchForkListing(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		slog.Debug("check update response 204 (current version is up to date)")
		return nil, nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("check update error %d - %.96s", resp.StatusCode, string(body))
	}
	return body, nil
}
