Test build of the thinking-budget work — **not** an official Ollama release, and not endorsed by the Ollama project. It exists so people can try the feature and report what breaks.

## New in this build

**Rebased onto Ollama `v0.34.4`, with llama.cpp `b11081`.** Take the runtime
archive as well as the binary (`ollama-linux-amd64-runtime.tgz`,
`ollama-windows-amd64-runtime.zip`): the two llama.cpp patches this series
carries compile into `lib/ollama`, and they are rebuilt here against the new
llama.cpp pin.

- **Thinking budgets follow 0.34.4's model-defined levels.** Upstream moved
  thinking levels onto the model (a model states the levels it supports and a
  default). The budget is ported onto that design: `"think": 1500` is still
  a token budget, and a level name is still a share of the output allowance.
  `minimal` is no longer raised to `low`, and `xhigh` means `max`. `--think`
  passes a level name it does not know to the server instead of refusing it.
  An integer ≤ 0 or a fraction is still refused.
- **Fixed: a budget was dropped on every model that states its levels.**
  0.34.4 resolves the requested level before the budget was read, so on
  qwen3.x, gemma4 and every other model with stated controls, a budget was
  replaced by the default level's share: `"think": 1500` reached the runner as
  1024 on a 4096 window. The budget is now read from the request.
- **Gemma 4: a tool call whose string value swallowed the next argument's
  name is rejected**, not executed. The shape is a string ending in
  `,<name>:` where `<name>` is not a key of the same object. It is rejected,
  never repaired, because repairing a cut value writes the fragment.

**Everything else is carried unchanged from `0.34.2-2-thinkbudget`**: the
response-scoped reasoning budget, the line-boundary close, Gemma 4 E2B/E4B
assistant drafters, and the Gemma 4, Qwen 3.5, LFM2 and qwen3-coder parser
fixes. The full list, with every patch's branch and commit, is `PATCHES.json`
on the `think-budget` branch.

`go test ./...` passes in full on this tree. The `cmd/launch` test that failed
on the stock tag under a tmpfs temp dir is fixed here (`up-codex-request-count-mtime`).

## Updating (unchanged since 0.34.2)

The 0.34.2 builds shipped the desktop app with three changes, on the
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

So these builds carry **`scripts/thinkbudget-install.ps1`**, which claims an
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
