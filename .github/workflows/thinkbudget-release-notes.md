Test build of the thinking-budget work — **not** an official Ollama release, and not endorsed by the Ollama project. It exists so people can try the feature and report what breaks.

## What it adds

`think` accepts a token count or an effort level, and a model can carry its own default:

| request | effect |
| --- | --- |
| `"think": 8192` | cap thinking at 8192 tokens |
| `"think": "minimal" \| "low" \| "medium" \| "high" \| "max"` | 1/16, 1/8, 1/4, 1/2, 4/5 of the request's context |
| `"think": true` | unrestricted, exactly as today |
| `PARAMETER think_budget 8192` / `high` | the model's own default |
| `PARAMETER think_budget_message "..."` | text written in just before the closing tag is forced |
| `"options": {"think_budget_message": "..."}` | same, per request |

The cap is enforced by llama.cpp's reasoning-budget sampler, not by trimming output: when the budget runs out the closing tag is forced, so the model finishes its answer instead of being cut off. The optional message tells the model *why* the block is closing, which on some models is the difference between a clean answer and the reasoning continuing inside the answer.

## Installing

These are the `ollama` binary only. The runtime it needs (llama.cpp **b10091**) is what stock **0.32.5** already ships, so:

1. Install official Ollama **0.32.5** normally.
2. Stop it (quit the tray app / `systemctl stop ollama`).
3. Replace the binary with the one from this release:
   - **Windows** — `%LOCALAPPDATA%\Programs\Ollama\ollama.exe`
   - **Linux** — `/usr/local/bin/ollama` (or wherever `which ollama` points)
   - **macOS** — inside `Ollama.app`, or your Homebrew/manual install path
4. Start it again. `ollama --version` should report `0.32.5-thinkbudget`.

Keep a copy of the original binary — reverting is just putting it back.

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
- macOS arm64 only; no Intel build.
- Only models with a thinking block are affected. Everything else is untouched.
- MLX runners ignore the fields.

Feedback in the pull request please: https://github.com/mann1x/ollama/pull/1
