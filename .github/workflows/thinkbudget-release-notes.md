Test build of the thinking-budget work — **not** an official Ollama release, and not endorsed by the Ollama project. It exists so people can try the feature and report what breaks.

## New in this build

**Rebased onto Ollama `v0.40.13`** (was `v0.40.12`). llama.cpp moves from
`b11901` to `b11963`, so the runtime archives are rebuilt; the ones from
`0.40.12-thinkbudget` do not match this build. Upstream also added a llama.cpp
patch that works around a crash in Metal's unary kernels; it only affects
Apple GPUs and is carried here as is. MLX and MLX-C are unchanged, and the MLX
archives are upstream's `v0.40.13` ones. Every carried patch applies to the
new base with the same changes it made before; none needed porting.

## Also in this build, from 0.40.12-thinkbudget

**Rebased onto Ollama `v0.40.12`** (was `v0.40.11`). llama.cpp moves from
`b11845` to `b11901`, so the runtime archives are rebuilt; the ones from
`0.40.11-thinkbudget` do not match this build. MLX and MLX-C move too
(upstream's prompt speed-up for Gemma, Flint and Laguna), and the MLX archives
are upstream's `v0.40.12` ones. Every carried patch applies to the new base
with the same changes it made before; none needed porting. Upstream's new LFM3
parser is built on the Qwen 3.5 one, so it inherits the Qwen 3.5 thinking and
tool-call tags this build carries.

## Also in this build, from 0.40.11-thinkbudget

**Rebased onto Ollama `v0.40.11`** (was `v0.40.10`). llama.cpp moves from
`b11803` to `b11845`, so the runtime archives are rebuilt; MLX is unchanged.

**qwen3-coder tool calls: a literal `</function>` inside a value, and
parameters holding parameters.** Upstream's #19228 keeps a `</function>`
that a parameter's value contains, such as `text = "</function>"`, as text
instead of ending the call there. That is carried here too, with one
difference: when the parameter is never closed after it, the model dropped
`</parameter>` and this build still recovers the call, as it did before.
Upstream also stopped accepting calls with a missing closing tag; this build
keeps recovering them. One case is now refused, however: a parameter that holds
other parameters (an `edits` list of `oldText`/`newText`, with the outer
parameter never closed) used to come out as a call with an empty `edits` and
two arguments the model never wrote. It now comes back as text, unrecovered.

## Also in this build, from 0.40.10-thinkbudget

Rebased onto Ollama `v0.40.10`: upstream's one change updated the Gemma
renderers, and no carried patch needed porting.

## Also in this build, from 0.40.9-1-thinkbudget

**A thinking budget works again on models whose template opens the thinking
block.** Many templates, qwen3's and gemma4's among them, end the prompt with
the opening `<think>` tag, so the model never writes that tag itself. Ollama
tells llama-server where the thinking block starts so the budget can start
counting there, and llama.cpp `b11669` stopped reading that part of the
request. From `0.40.6-thinkbudget` through `0.40.9-thinkbudget`, a
`think_budget` on such a model was reported in the response but never
enforced: the model thought until it ran out of output tokens, often with no
answer. This build carries a llama.cpp patch that reads it again. The runtime
archives are rebuilt for it, so the ones from `0.40.9-thinkbudget` do not
match this build; the MLX archives are unchanged.

## Also in this build, from 0.40.9-thinkbudget

**Rebased onto Ollama `v0.40.9`, with llama.cpp `b11803`** (was `v0.40.6` and
`b11669`; `v0.40.7` and `v0.40.8` are skipped). Every carried patch applies to
the new base with the same changes it made before. Two needed porting onto
upstream's own changes: the 400 for an unreadable `think_budget` now sits in
the error mapping upstream split out of the request handlers, and one test was
updated for a scheduler field upstream made atomic. The runtime archives are
rebuilt for the new llama.cpp; the ones from `0.40.6-thinkbudget` do not match
this build.

**Thinking goes to the chat template only on assistant turns.** On models run
through llama-server's own chat template, a message's thinking was passed to
the template whatever its role. Only an assistant turn carries reasoning, but
a client can leave thinking on a user or tool message (the OpenAI-compatible
endpoint maps `reasoning` on any message into it), and a template that does
not check the role itself would then render it. Ollama's own renderers already
wrote thinking for assistant turns only; this path now does the same.

**The MLX archives no longer carry the CUDA libraries Ollama shares with
llama.cpp.** From 0.40.9, upstream keeps one copy of cuBLAS, cuBLASLt, the CUDA
runtime, NVRTC and nvJitLink, in the main archive's `lib/ollama/cuda_v13`, and
MLX loads them from there. The MLX archives on this release are upstream's
0.40.9 files, unchanged, so they need that folder next to them: extract them
over an Ollama 0.40.9 install, whose `cuda_v13` has those libraries. On
Windows, upstream's `mlx.dll` is now built to find cuDNN next to itself and
CUDA in `../cuda_v13`, instead of in the folders upstream's build machine had
them in, so this release has no `-reldir` zip: use
`ollama-windows-amd64-mlx.zip`.

**Also in this build, from the 0.40.6 series:**

**`ollama ps` reports a model's real size again.** Ollama works out how much
memory a loaded model uses by reading the buffer sizes llama-server prints as
it loads. Two things threw that count off. A line could arrive in two pieces,
because llama-server's output is read in chunks that do not end at line ends,
and a line that arrived in two pieces was not counted at all. And with a
speculative draft (an MTP model, or a separate draft model), the draft's
buffers carry the same names as the main model's, so they replaced them
instead of adding to them. A Qwen3.5 9B Q6_K with an MTP draft, fully on the
GPU, showed as 0.37 GB, then 0.88 GB after its first request; it now shows
9.93 GB, the size llama-server itself projects. The scheduler places models
by this number, so a model that looked small could be given room it did not
have.

**Rebased onto Ollama `v0.40.6`, with llama.cpp `b11669`** (was `v0.40.4` and
`b11521`; `v0.40.5` is skipped). Every carried patch applies to the new base
with the same changes it made before. Upstream moved MLX to `bebd6422`;
MLX-C is unchanged. The runtime archives are rebuilt for the new llama.cpp;
the ones from `0.40.4-thinkbudget` do not match this build.

**Gemma 4 no longer leaks a channel tag into its reasoning.** When Gemma 4
opened its thought channel again while it was already thinking, the tag and
the channel's name (`<|channel>thought`) showed up in the reasoning text. The
parser now drops them and continues the reasoning on a new line.

**A model pulled for both runners runs on MLX when asked to.** With the GGUF
and the MLX copy of a model both downloaded, a request with `runner: "mlx"`
started the MLX runner on the GGUF copy and failed with `config.json not
found in manifest` (Windows and Linux; Apple silicon picked MLX already). The
MLX runner now always opens the MLX copy.

**Also in this build, from the 0.40.4 series:**

**An unreadable `think_budget` is refused with 400, not 500.** A word that
names no share (`"think_budget": "bogus"`), a fraction, zero, a negative
number or a value that is neither a number nor a string was already refused,
with the reason, but answered as a server error. It is the client's error, so
`/api/chat` and `/api/generate` now answer 400, with the same message.

**Also in this build, from the 0.40.3 series:**

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

**MLX runtimes are on this release, for every platform**, with
`mlx-runtime.txt` listing each file's sha256. Safetensors models run on MLX.
`ollama-linux-amd64-mlx.tar.zst`, `ollama-windows-amd64-mlx.zip` and
`ollama-darwin-mlx.tgz` are upstream's 0.40.9 bytes, unchanged.

**Everything else is carried unchanged**: the response-scoped reasoning
budget, the line-boundary close, Gemma 4 E2B/E4B assistant drafters, the
Modelfile round-trip fixes, and the Gemma 4, Qwen 3.5, LFM2 and qwen3-coder
parser fixes. The full list, with every patch's branch and commit, is
`PATCHES.json` on the `think-budget` branch: 29 patches on the `v0.40.9` tag.

`go test ./...` passes in full on this tree. The carried llama.cpp patches
apply in full to a fresh `b11803` checkout, and llama.cpp's own
`test-reasoning-budget` passes against it.

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
