WAILS_CLI ?= go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

.PHONY: harness-check harness-verify harness-review-gate frontend-install frontend-build desktop-shell-build desktop-shell-dev

harness-check:
	bash scripts/harness/check.sh

harness-verify: harness-check

harness-review-gate:
	@if [ -z "$(PLAN)" ]; then echo "usage: make harness-review-gate PLAN=path/to/plan.md" >&2; exit 2; fi
	bash scripts/harness/review_gate.sh --plan "$(PLAN)"

frontend-install:
	npm --prefix frontend install

frontend-build:
	npm --prefix frontend run build

desktop-shell-build:
	$(WAILS_CLI) build -clean

desktop-shell-dev:
	$(WAILS_CLI) dev
