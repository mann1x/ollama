Test build of the thinking-budget work — **not** an official Ollama release, and not endorsed by the Ollama project. It exists so people can try the feature and report what breaks.

## New in this build

**The thinking budget bounds a response, not a block.** A budget that re-armed in full on every thinking block bounded a block, not a turn: a model that closes each block by itself just short of its window and opens another was never cut. Measured on gemma4 through a coding agent — six consecutive blocks against an 8,000-token budget, none exhausted, a 32,000-token output cap consumed, and a turn that produced neither an answer nor a tool call. The budget is now spent across the response, and a tool call forgives what the thinking before it spent, so a long agentic turn does not run out of thinking after its first few steps.

**The cut lands at the end of a line.** The forced message used to be spliced in wherever the token counter ran out, mid word: `Actually, I'Considering the limited time by the user...`. It now waits for the model to finish the line, and gives up after 64 tokens if no newline arrives — base64, a long single-line table, a run-on paragraph.

Both live in llama.cpp's reasoning-budget sampler, which is why this release also ships a runtime — see Installing.

## What it adds

`think` accepts a token count or an effort level, and a model can carry its own default:

| request | effect |
| --- | --- |
| `"think": 8192` | cap thinking at 8192 tokens |
| `"think": "minimal" \| "low" \| "medium" \| "high" \| "max"` | 1/16, 1/8, 1/4, 1/2, 4/5 of the response the request allows — `num_predict` when it sets one, the context length otherwise |
| `"think": true` | unrestricted, exactly as today |
| `PARAMETER think_budget 8192` / `high` | the model's own default |
| `PARAMETER think_budget_message "..."` | text written in just before the closing tag is forced |
| `"options": {"think_budget_message": "..."}` | same, per request |

The cap is enforced by llama.cpp's reasoning-budget sampler, not by trimming output: when the budget runs out the closing tag is forced, so the model finishes its answer instead of being cut off. The optional message tells the model *why* the block is closing, which on some models is the difference between a clean answer and the reasoning continuing inside the answer.

## Also in this build

Three fixes found by running the budget under a real coding agent. Each is independent of the budget and helps any tool-using model.

| fix | what went wrong before |
| --- | --- |
| [#3](https://github.com/mann1x/ollama/pull/3) repeat guard | A generation was aborted after 31 identical tokens and the abort was reported as success, so the stream ended with no `done` and every client raised "Did not receive done or success response in stream". Base64 of a file, a hex dump, or a run of indentation was enough to trigger it. The run is now measured in characters with a budget no real payload reaches, a repeating unit of up to 32 tokens is recognised, and a generation stopped this way ends with `done_reason: "repeat"`. |
| [#4](https://github.com/mann1x/ollama/pull/4) truncated tool calls | A response that ran out of tokens mid-tool-call handed over the arguments that had arrived, so a caller saw a complete-looking call with a required argument missing and reported a schema error naming a field the model was still writing. Such a call is now dropped, and only when the generation actually stopped at the limit. |
| [#5](https://github.com/mann1x/ollama/pull/5) gemma4 tool calls | A Gemma 4 tool call that was complete apart from its final `}` failed to parse and was dropped silently, so the caller received an empty response. It is now recovered — but only when the model emitted its closing tag, which rules out truncation. |

## Installing

Earlier builds were the `ollama` binary alone, because the runtime the budget needed was llama.cpp **b10091** — exactly what stock **0.32.5** ships. That is no longer true: the two changes at the top of these notes are in the budget sampler itself, which compiles into `lib/ollama`, not into `ollama.exe`. On Windows, install both or you get the half that asks for behaviour the runtime does not have.

1. Install official Ollama **0.32.5** normally.
2. Stop it (quit the tray app / `systemctl stop ollama`).
3. Replace the binary with the one from this release:
   - **Windows** — `%LOCALAPPDATA%\Programs\Ollama\ollama.exe`
   - **Linux** — `/usr/local/bin/ollama` (or wherever `which ollama` points)
   - **macOS** — inside `Ollama.app`, or your Homebrew/manual install path
4. **Windows only:** unpack `ollama-windows-amd64-runtime.zip` over `%LOCALAPPDATA%\Programs\Ollama\lib\ollama`, replacing the files it contains. It holds the base runtime — `llama-server.exe`, `libllama-common.dll`, `libllama.dll`, the `ggml-cpu-*` variants. Leave the `cuda_v12\`, `cuda_v13\`, `rocm_v7_1\` and `vulkan\` folders alone: the change is in the base set, and the backends reach it through ggml's C ABI, so your GPU acceleration is untouched.
5. Start it again. `ollama --version` should report `0.32.5-thinkbudget`.

Keep a copy of the original binary and of the files you replace in `lib\ollama` — reverting is just putting them back.

**Linux and macOS** get the binary only for now, so the budget works as it did in earlier builds (per-block, cut wherever the counter lands). Build the runtime from the branch if you want the two new behaviours there.

## Trying it

```bash
curl http://localhost:11434/api/chat -d '{
  "model": "your-thinking-model",
  "messages": [{"role":"user","content":"a hard question"}],
  "think": "medium",
  "options": {"num_ctx": 32768}
}'
```

Or bake it into a model:

```
FROM your-thinking-model
PARAMETER think_budget medium
PARAMETER think_budget_message """

OK, I have enough to answer now.
"""
```

A client that sends no `think` field still gets the model's own budget, which is the point of the Modelfile form — coding agents generally do not send one.

## Caveats

- Unsigned, built by GitHub Actions from this fork. Windows SmartScreen will complain.
- macOS arm64 only; no Intel build. Gatekeeper blocks unsigned downloads — `xattr -d com.apple.quarantine ollama-darwin-arm64` before running it.
- The Linux binary is built against glibc 2.28, so it runs on RHEL 8, Ubuntu 20.04 and Debian 11 upwards. Verify a download against `sha256sum.txt` before replacing anything.
- The Windows runtime archive is built with the MSYS2 clang64 toolchain, which is what the shipped DLLs use, so it drops in beside the CUDA and Vulkan backends already installed. It replaces the CPU/base set only.
- Only models with a thinking block are affected. Everything else is untouched.
- MLX runners ignore the fields.

Feedback in the pull requests please: [#1](https://github.com/mann1x/ollama/pull/1) for the budget, [#3](https://github.com/mann1x/ollama/pull/3) / [#4](https://github.com/mann1x/ollama/pull/4) / [#5](https://github.com/mann1x/ollama/pull/5) for the fixes above.
