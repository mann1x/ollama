//go:build windows || darwin

package updater

import (
	"log/slog"
	"os"
)

// Whether this machine has consented to being upgraded.
//
// Set by the app once the settings store is open, and nil before that. Nil
// means "nobody has said", which is treated as consent -- it is upstream's
// behaviour, and a build with no settings at all should still be able to
// update itself.
var AutoUpdateAllowed func() (bool, error)

// Whether a staged installer may be applied, and whether it is worth keeping.
//
// The download is already gated on the setting (updater.go), but nothing was
// gating the *bundle*: turning auto-update off cancels a download in flight
// and leaves a completed one on disk forever, where IsUpdatePending keeps
// finding it, the tray keeps offering "Restart to update", and every path that
// reaches DoUpgrade applies an update this machine has declined. Measured on
// pandorum 2026-09-12, whose upgrade.log records a silent install at 12:17
// from a bundle staged before the setting was turned off.
func stagedUpdateConsent() (allowed, keep bool) {
	if AutoUpdateAllowed == nil {
		return true, true
	}
	enabled, err := AutoUpdateAllowed()
	if err != nil {
		// Settings that cannot be read are not permission. Upstream warns and
		// upgrades anyway, which turns one unreadable database into a silent
		// replacement of the running build -- but an unreadable setting is not
		// a decision to throw the download away either, so the bundle stays.
		slog.Warn("cannot read the auto-update setting; not applying the staged update", "error", err)
		return false, true
	}
	if !enabled {
		return false, false
	}
	return true, true
}

// Apply the consent decision to the bundles found on disk.
//
// Returns the bundle to use, or "" when there is none this machine will take.
// Discarding is the point rather than a tidy-up: a bundle left in place is
// re-offered on every startup, and the next time the setting is read as
// enabled -- or fails to be read at all -- it is applied.
func consentedStagedUpdate(files []string) string {
	if len(files) == 0 {
		return ""
	}
	allowed, keep := stagedUpdateConsent()
	if allowed {
		if len(files) > 1 {
			// Shouldn't happen
			slog.Warn("multiple update downloads found, using first one", "bundles", files)
		}
		return files[0]
	}
	if keep {
		slog.Debug("leaving the staged update in place", "bundles", files)
		return ""
	}
	for _, f := range files {
		slog.Info("auto-update is disabled; discarding the staged update", "bundle", f)
		if err := os.Remove(f); err != nil {
			slog.Warn("failed to discard the staged update", "bundle", f, "error", err)
		}
	}
	return ""
}
