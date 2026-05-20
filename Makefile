WAILS_CLI ?= go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

.PHONY: harness-check harness-verify harness-review-gate frontend-install frontend-typecheck frontend-test-unit frontend-build fix-generated-bindings desktop-shell-build desktop-shell-dev alpha-shell-smoke alpha-shell-smoke-launch worktree-whitespace-check alpha-shell-verify

harness-check:
	bash scripts/harness/check.sh

harness-verify: harness-check

harness-review-gate:
	@if [ -z "$(PLAN)" ]; then echo "usage: make harness-review-gate PLAN=path/to/plan.md" >&2; exit 2; fi
	bash scripts/harness/review_gate.sh --plan "$(PLAN)"

frontend-install:
	npm --prefix frontend install

frontend-typecheck:
	npm --prefix frontend run typecheck

frontend-test-unit:
	npm --prefix frontend run test:unit

frontend-build:
	npm --prefix frontend run build

fix-generated-bindings:
	@if [ -f frontend/wailsjs/go/models.ts ]; then perl -0pi -e 's/[ \t]+$$//mg; s/\n+\z/\n/' frontend/wailsjs/go/models.ts; fi

desktop-shell-build:
	$(WAILS_CLI) build -clean
	$(MAKE) fix-generated-bindings

desktop-shell-dev:
	$(WAILS_CLI) dev

alpha-shell-smoke: desktop-shell-build
	bash scripts/harness/alpha_shell_smoke.sh --check-build

alpha-shell-smoke-launch: desktop-shell-build
	bash scripts/harness/alpha_shell_smoke.sh --launch

worktree-whitespace-check:
	bash scripts/harness/check_worktree_whitespace.sh

alpha-shell-verify: harness-verify
	go test ./...
	$(MAKE) frontend-typecheck
	$(MAKE) frontend-test-unit
	$(MAKE) frontend-build
	$(MAKE) alpha-shell-smoke
	$(MAKE) worktree-whitespace-check
