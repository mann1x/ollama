//go:build (windows || darwin) && updater_live

package updater

import (
	"context"
	"testing"
	"time"

	"github.com/ollama/ollama/app/version"
)

// TestLiveAppUpdate exercises the real update endpoint this build ships
// pointing at. It is excluded from normal test runs because it depends on the
// network.
//
// It no longer downloads anything, and the reason is the change it is here to
// guard. Upstream's version of this test spoofed an old version, asked
// ollama.com, and asserted that an installer came back; this build asks the
// fork it is released from, and that series publishes a binary and the base
// runtime rather than an installer -- so the correct answer to "is there an
// update" is "no", and there is nothing to stage. What is still worth checking
// live is everything up to that point: that the endpoint is reachable, that
// its answer parses, and that a release far newer than anything published is
// still declined for want of something installable.
//
// Run with:
//
//	go test -tags updater_live -run TestLiveAppUpdate ./app/updater
func TestLiveAppUpdate(t *testing.T) {
	const spoofedVersion = "0.20.0"

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	oldVersion := version.Version
	t.Cleanup(func() { version.Version = oldVersion })
	version.Version = spoofedVersion

	if UpdateCheckURLBase == "" {
		t.Fatal("this build has update checks compiled off")
	}
	t.Logf("update endpoint %s", UpdateCheckURLBase)

	updater := &Updater{}
	available, updateResp := updater.checkForUpdate(ctx)
	if available {
		// Not a failure of the endpoint -- a release that carries an installer
		// is a real change in what this series publishes, and the deployment
		// notes say it does not. Fail loudly so the claim gets revisited.
		t.Fatalf("the fork offered an installable update: version=%q url=%q", updateResp.UpdateVersion, updateResp.UpdateURL)
	}
}

// TestLiveForkListingParses proves the answer is a listing this build
// understands, rather than an error page that happens to decline.
func TestLiveForkListingParses(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	req, err := newForkRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	body, err := fetchForkListing(req)
	if err != nil {
		t.Fatal(err)
	}
	releases, err := parseForkListing(body)
	if err != nil {
		t.Fatalf("the update endpoint did not return a releases listing: %v", err)
	}
	if len(releases) == 0 {
		t.Fatal("the update endpoint returned no releases at all")
	}
	t.Logf("newest release %q with %d assets", releases[0].TagName, len(releases[0].Assets))
}
