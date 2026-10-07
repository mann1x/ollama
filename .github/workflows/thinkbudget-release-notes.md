Test build of the thinking-budget work — **not** an official Ollama release, and not endorsed by the Ollama project. It exists so people can try the feature and report what breaks.

## New in this build

**Rebased onto Ollama `v0.40.0`, with llama.cpp `b11351`.** Take the runtime
archive as well as the binary (`ollama-linux-amd64-runtime.tgz`,
`ollama-linux-arm64-runtime.tgz`, `ollama-windows-amd64-runtime.zip`): the llama.cpp patches this series carries
compile into `lib/ollama`, and they are rebuilt here against the new llama.cpp
pin. A runtime from `0.35.1-thinkbudget` is llama.cpp `b11232` and does not
match this binary.

What 0.40.0 itself brings is upstream's, unchanged here: runner-specific
manifests under one tag, with older Ollama GGUFs converted on first load to a
form llama.cpp reads without Ollama's compatibility patch; multimodal
embeddings; decision models on MLX (System One); and image positions kept in
OpenAI tool results. Upstream describes the manifest change as preparation for
removing that compatibility patch. This build still carries it, with the two
llama.cpp patches of this series on top.

**Upstream's MLX runtimes are on this release, for every platform**, bytes
unchanged: `ollama-linux-amd64-mlx.tar.zst`, `ollama-windows-amd64-mlx.zip`
and `ollama-darwin-mlx.tgz`, with `mlx-runtime.txt` listing each file's
sha256. Safetensors models run on MLX.

**Everything else is carried unchanged from `0.35.1-thinkbudget`**: thinking
budgets on the model-defined levels, the response-scoped reasoning budget, the
line-boundary close, Gemma 4 E2B/E4B assistant drafters, the Modelfile
round-trip fixes, and the Gemma 4, Qwen 3.5, LFM2 and qwen3-coder parser fixes.
The full list, with every patch's branch and commit, is `PATCHES.json` on the
`think-budget` branch: 24 patches, each rebased onto the `v0.40.0` tag.

`go test ./...` passes in full on this tree, and llama.cpp's own
`test-reasoning-budget` passes against `b11351` with the carried patches
applied.

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
