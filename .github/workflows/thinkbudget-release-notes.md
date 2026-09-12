Test build of the thinking-budget work — **not** an official Ollama release, and not endorsed by the Ollama project. It exists so people can try the feature and report what breaks.

## New in this build

**Nothing new in the thinking-budget work itself.** This is the same feature
set as the 0.33.3 test build, rebased onto Ollama **0.34.0** so it can be
installed over the current release rather than over one you have to hold back.

One test changed, and it is worth naming because it marks where this series
parts company with upstream. `TestFromResponsesRequest_ReasoningEffort` probes
the invalid-value path with `think: 3`, which upstream is right to refuse — its
`ThinkValue.IsValid` has no case for integers at all, so every one of them
falls through to `default`. This series gives an integer a meaning: a
thinking-token budget, valid when positive. So `3` became legal, a very small
budget but nothing about it malformed, and the case stopped testing what it
names. The probe moved to `0`, which an integer still cannot be.

## What changed about updating

Earlier builds in this series replaced `ollama` and the runtime and left the
desktop app alone — and the stock app checks `ollama.com` hourly, is offered
the official release, and installs it, putting the stock binary back over this
one. That is not hypothetical: on one machine here the check offered v0.34.0
from 09:10, the installer ran at 12:12, and the server it left behind did not
start at all.

Turning **automatic updates off** in the app's settings was the documented
workaround, and it does not work either. Disabling cancels a download that is
in flight; a download that already finished stays staged on disk, nothing ever
clears it, and every path that applies it — the tray's "Restart to update", the
upgrade at startup — went ahead without consulting the setting. A machine with
the setting off was silently upgraded here on 2026-09-12.

So this release ships the desktop app (Windows) with three changes. It is
attached as **`ollama-app-windows-amd64.exe`** and must be renamed to
**`ollama app.exe`** when you copy it in -- GitHub rewrites spaces in asset
names, so it cannot be attached under the name it has to be installed under:

- The update check asks **this fork's releases**, not `ollama.com`. Since these
  releases carry a binary and the runtime rather than an installer, the honest
  answer today is always "no update" — which is the intended resting state, not
  a failure. Point it somewhere else, or switch it off entirely, by rebuilding
  with a different `UpdateCheckURLBase`.
- A staged installer is **discarded** when automatic updates are off, so the
  setting also clears what was downloaded before you turned it off.
- Settings that cannot be read are no longer treated as permission. Upstream
  logs a warning and upgrades anyway; this build declines and leaves the
  download in place.

## Known limits of this build

`ollama app.exe` here is **unsigned**, so Windows SmartScreen will warn the
first time you run it. This release contains no installer — copy the files over
an existing install.

On **macOS** only the binary and runtime are shipped; the stock app is
unchanged and will still replace this build. Turn automatic updates off there,
and be aware of the staging bug above: if an update had already downloaded, the
setting will not stop it.
