Test build of the thinking-budget work — **not** an official Ollama release, and not endorsed by the Ollama project. It exists so people can try the feature and report what breaks.

## New in this build

**Agent steps with thinking off no longer re-read the previous step.** With
thinking off, the Qwen3.5 renderer (also used for tev1 and the Qwen3.6 tags)
starts each reply after an empty think block. The history then rendered that
reply without the block, so the next request's prompt changed right after
`assistant`. Qwen3.5 and 3.6 can only resume their cache at a saved
checkpoint before that point, so every tool-calling step evaluated the whole
previous step again. A reply after the last user message is now rendered
exactly as it was generated. Measured through xOllama on qwen3.5:2b, thinking
off, three agent steps: step 2 reuses 757 of 778 prompt tokens (315 before),
step 3 1,282 of 1,303 (770 before). Gemma 4's large renderer had the same
mismatch with its empty thought channel and gets the same rule. Earlier turns,
before the last user message, keep the model's template form.

**Also in this build, from v0.40.3-thinkbudget.1:**

**A thinking level is no longer a thinking budget.** Until now `think` did two
jobs: a level such as `"high"` was handed to the model as its reasoning level
and was also read as a cap on thinking (half the response for `"high"`), and
that cap overruled `think_budget`. A number in `think` was a cap too. They are
two settings now:

- **`think`** is what it is in stock Ollama: `true`, `false`, or one of the
  model's own levels, handed to the model as written. It never caps anything.
  A number in `think` is refused, as stock refuses it. `reasoning_effort` on
  the OpenAI-compatible route is stock again as well (`minimal` is `low`).
- **`think_budget`** is the only cap: a token count, or `minimal`, `low`,
  `medium`, `high` or `max` as a share of the response (a sixteenth, an eighth,
  a quarter, a half, four fifths). Set it per request in `options`, or per
  model with `PARAMETER think_budget`. A request's value replaces the model's.

What changes for you:

- **Models that carry `PARAMETER think_budget` behave as before** when the
  request sends `think: true` or no `think` at all.
- **A request that sent a level to get a cap gets no cap from it any more.**
  Send `options.think_budget` instead. Example: `"think": "high"` used to cap
  thinking at half the response; now send `"think": true, "options":
  {"think_budget": "high"}` for the same cap, or keep `"think": "high"` beside
  it if the model has that level.
- **A level and a budget together both apply.** `"think": "low"` with
  `"think_budget": 8000` asks the model for its low level and caps thinking at
  8,000 tokens. Before, the level won and the 8,000 was ignored.
- **`think: 4096` is now an error.** Use `options.think_budget: 4096`.
- **A `think_budget` word that names no share is refused** instead of being
  accepted and capping nothing. A count written as a string (`"4096"`) is read
  as a count.
- **`ollama run --think=<n>`** no longer takes a number; `--think` is stock.
  Use `/set parameter think_budget <n>`.
- **Anthropic-compatible requests:** `thinking.budget_tokens` still caps
  thinking; it now travels as `think_budget`.
- `think_budget` and `think_budget_tokens` in responses and in `/api/show`
  report the budget from `think_budget` only. `/api/show` still takes `think`,
  but only `false` changes its answer.

**Same base as `0.40.3-thinkbudget`:** Ollama `v0.40.3`, llama.cpp `b11351`.
Only the `ollama` binaries are new. The runtime archives
(`ollama-linux-amd64-runtime.tgz`, `ollama-linux-arm64-runtime.tgz`,
`ollama-windows-amd64-runtime.zip`) are that release's files, byte for byte,
with the same sha256. If you already have them, keep them.

**MLX runtimes are on this release, for every platform**, with
`mlx-runtime.txt` listing each file's sha256. Safetensors models run on MLX.
`ollama-linux-amd64-mlx.tar.zst`, `ollama-windows-amd64-mlx.zip` and
`ollama-darwin-mlx.tgz` are upstream's 0.40.3 bytes, unchanged.
`ollama-windows-amd64-mlx-reldir.zip` is the file from `0.40.1-thinkbudget`
(MLX `a59cc231`, which 0.40.3 still uses): built here from the same MLX and
CMake as upstream's, with one difference. Upstream's `mlx.dll` looks for cuDNN
in `C:/Program Files/NVIDIA/CUDNN/bin/x64`, the folder its build machine had it
in, and panics on the first generation that uses cuDNN on any machine without
it there. Ours looks next to `mlx.dll`, where the zip puts the cuDNN and CUDA
DLLs. On Windows, use the `-reldir` zip.

**Everything else is carried unchanged from `0.40.3-thinkbudget`**: the
response-scoped reasoning budget, the line-boundary close, Gemma 4 E2B/E4B
assistant drafters, the Modelfile round-trip fixes, and the Gemma 4, Qwen 3.5,
LFM2 and qwen3-coder parser fixes. The full list, with every patch's branch and
commit, is `PATCHES.json` on the `think-budget` branch: 24 patches on the
`v0.40.3` tag.

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
