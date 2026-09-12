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

## Known limits of this build

It replaces `ollama` and the base runtime. It does **not** replace the desktop
app, so on Windows and macOS the stock app's updater keeps checking
ollama.com on its own hourly schedule, and an update it installs will put the
stock binary back over this one. If you want this build to stay put, turn
**automatic updates off in the Ollama app's settings** before installing it.
That is a setting in the app, not something this release can carry.
