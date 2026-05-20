WAILS_CLI ?= go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

.PHONY: harness-check harness-verify harness-review-gate frontend-install frontend-build fix-generated-bindings desktop-shell-build desktop-shell-dev

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

fix-generated-bindings:
	@if [ -f frontend/wailsjs/go/models.ts ]; then perl -0pi -e 's/[ \t]+$$//mg; s/\n+\z/\n/' frontend/wailsjs/go/models.ts; fi

desktop-shell-build:
	$(WAILS_CLI) build -clean
	$(MAKE) fix-generated-bindings

desktop-shell-dev:
	$(WAILS_CLI) dev
