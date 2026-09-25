Test build of the thinking-budget work — **not** an official Ollama release, and not endorsed by the Ollama project. It exists so people can try the feature and report what breaks.

## New in this build

**Mostly the runtime.** Two of the four changes are llama.cpp patches, which
compile into `lib/ollama`, not into the `ollama` binary — so on Linux take
`ollama-linux-amd64-runtime.tgz` as well, and on Windows
`ollama-windows-amd64-runtime.zip`. The binary alone does not carry them.

- **A spent response budget now stays quiet.** The runtime of `0.34.2-1`
  still carried an older copy of the reasoning-budget patch. With a
  response-scope budget spent, a model that reopened its thinking block got the
  whole wrap-up message forced into it again, every time — measured through a
  coding agent at thirty-two identical copies in one turn, ending at the output
  cap with no answer. A reopened block is now closed with the end tag alone,
  and the start tag is barred while the allowance is gone; a reset sequence
  (a tool call) lifts both.
- **Gemma 4 E2B/E4B assistant drafters load.** `check_tensor_dims` read the
  drafter's deliberately unchecked `masked_embd_*` shapes as "must be a
  scalar", and the error path then threw
  `vector::_M_range_check: __n (which is 0) >= this->size() (which is 0)`
  while trying to print the mismatch. 12B, 26B-A4B and 31B were never
  affected. Acceptance stays below what the drafter was trained for, because
  its ordered-embedding head is not implemented; output is unaffected, since
  every drafted token is verified.
- **Gemma 4: a tool call the parser cannot read no longer vanishes.** It
  arrives as content, tags included, instead of an empty turn — deliberately
  not repaired, since the captured case was a degenerating model.
- **LFM2: `"think": false` keeps the reasoning block out of the answer.**
  LFM2.5's template has no switch to stop reasoning, so the block is now
  recognised and discarded rather than returned as the answer.

**Everything else is identical to the `0.34.2-1-thinkbudget` build**,
including the Gemma 4 fix for a tool call opened inside the thinking channel
that led that build.

One upstream test fails on this tag and is **not** something this build causes
or fixes: `cmd/launch`'s `TestCodexAppCountsOnlyOllamaRequestsInRegularProfile`
(`regular profile Ollama request count = 0, want 2`). It fails identically on
the pristine `v0.34.2` tag.

## What changed about updating — and what was still wrong

The previous build shipped the desktop app with three changes, on the
understanding that the app was what replaced these builds with stock ones. The
app *is* one of the things that does it, and those three changes stand:

- The update check asks **this fork's releases**, not `ollama.com`. Since these
  releases carry a binary and the runtime rather than an installer, the honest
  answer today is always "no update" — the intended resting state, not a
  failure.
- A staged installer is **discarded** when automatic updates are off, so the
  setting also clears what was downloaded before you turned it off.
- Settings that cannot be read are no longer treated as permission. Upstream
  logs a warning and upgrades anyway; this build declines.

**They were necessary and they were not sufficient.** On 2026-09-17 a machine
running this series was upgraded to stock 0.34.2 anyway, fifteen hours after
`0.34.0-thinkbudget` was installed on it — `ollama.exe` and `ollama app.exe`
replaced four seconds apart. Nothing in Ollama did it.

The **Microsoft Store** did, through the Windows Package Manager. Stock
Ollama's Inno Setup installer leaves an Add/Remove Programs entry, the winget
manifest for `Ollama.Ollama` carries no ProductCode, so the correlation is made
on DisplayName and Publisher — and that entry survives whatever binaries are
copied over the top. The Store then upgrades on its own schedule, in a separate
process that consults none of the three changes above.

What made it hard to see: a running server keeps the binary it has already
loaded. The upgraded machine went on answering `0.34.0-thinkbudget` for another
fourteen hours and only failed when it was restarted — at which point stock
0.34.2 did not start at all (`Failed to start: Unable to init instance`).

So this build adds **`scripts/thinkbudget-install.ps1`**, which claims an
Add/Remove Programs identity of its own — its own AppId GUID,
`Ollama think-budget`, publisher `mann1x` — so there is nothing left for the
Store to match. It backs up the stock key first, refuses to swap binaries under
a running process, and then asks winget whether it worked rather than assuming:
afterwards `winget list --id Ollama.Ollama` answers *"No installed package
found matching input criteria."*

If you are running an earlier build of this series on Windows, that script is
worth running on its own (`-IdentityOnly`) even if you do not take this build.

## Known limits of this build

`ollama app.exe` here is **unsigned**, so Windows SmartScreen will warn the
first time you run it. It is attached as **`ollama-app-windows-amd64.exe`** and
must be renamed to **`ollama app.exe`** when you copy it in — GitHub rewrites
spaces in asset names. This release contains no installer; copy the files over
an existing install.

On **macOS** only the binary and runtime are shipped; the stock app is
unchanged and will still replace this build.
