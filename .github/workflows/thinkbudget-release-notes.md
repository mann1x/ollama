Test build of the thinking-budget work — **not** an official Ollama release, and not endorsed by the Ollama project. It exists so people can try the feature and report what breaks.

## New in this build

**Rebased onto Ollama `v0.40.2`, still with llama.cpp `b11351`.** Only the
`ollama` binaries are new. Nothing the llama runtime is compiled from changed
since `0.40.1-thinkbudget`, so the runtime archives here
(`ollama-linux-amd64-runtime.tgz`, `ollama-linux-arm64-runtime.tgz`,
`ollama-windows-amd64-runtime.zip`) are that release's files, byte for byte,
with the same sha256. If you already have them, keep them.

What 0.40.2 itself brings is upstream's, unchanged here: `ollama list` no
longer shows a second entry for a model Ollama has converted, and
`ollama launch claude` uses the model's full context length.

**About the "model upgrades" in upstream's 0.40.2 notes.** That behaviour is
not new in 0.40.2: it arrived in 0.40.0 and is in every build since, this
series included. The first time a model stored in Ollama's older GGUF layout is
loaded, the server writes a second copy in a layout plain llama.cpp reads, in
the background, and uses that copy afterwards. The original stays on disk as
the copy an older Ollama can still read, so an affected model takes roughly
twice its size until the backup is removed. It is skipped when disk space is
short. Upstream's notes carry a script that deletes the backups.

**MLX runtimes are on this release, for every platform**, with
`mlx-runtime.txt` listing each file's sha256. Safetensors models run on MLX.
`ollama-linux-amd64-mlx.tar.zst`, `ollama-windows-amd64-mlx.zip` and
`ollama-darwin-mlx.tgz` are upstream's 0.40.2 bytes, unchanged.
`ollama-windows-amd64-mlx-reldir.zip` is the file from `0.40.1-thinkbudget`
(MLX `a59cc231`, which 0.40.2 still uses): built here from the same MLX and
CMake as upstream's, with one difference. Upstream's `mlx.dll` looks for cuDNN
in `C:/Program Files/NVIDIA/CUDNN/bin/x64`, the folder its build machine had it
in, and panics on the first generation that uses cuDNN on any machine without
it there. Ours looks next to `mlx.dll`, where the zip puts the cuDNN and CUDA
DLLs. On Windows, use the `-reldir` zip.

**Everything else is carried unchanged from `0.40.1-thinkbudget`**: thinking
budgets on the model-defined levels, the response-scoped reasoning budget, the
line-boundary close, Gemma 4 E2B/E4B assistant drafters, the Modelfile
round-trip fixes, and the Gemma 4, Qwen 3.5, LFM2 and qwen3-coder parser fixes.
The full list, with every patch's branch and commit, is `PATCHES.json` on the
`think-budget` branch: 24 patches, each rebased onto the `v0.40.2` tag.

`go test ./...` passes in full on this tree. The llama.cpp sources and the
carried patches are byte-identical to `0.40.0-thinkbudget`'s, where
llama.cpp's own `test-reasoning-budget` passes against `b11351`.

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
