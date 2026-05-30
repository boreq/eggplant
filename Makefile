.PHONY: ci
ci: ci-backend ci-frontend

.PHONY: dev
dev:
	./tools/dev.sh $(CONFIG)

.PHONY: dev-public
dev-public:
	./tools/dev.sh --public $(CONFIG)

.PHONY: ci-backend
ci-backend:
	$(MAKE) -C backend ci

.PHONY: ci-frontend
ci-frontend:
	$(MAKE) -C frontend ci

.PHONY: build
build:
	version=$$($(MAKE) --no-print-directory version); \
	$(MAKE) -C frontend install; \
	$(MAKE) -C frontend build VUE_APP_VERSION=$$version; \
	cp -r frontend/dist/. backend/entrypoints/http/frontend/; \
	$(MAKE) -C backend build-with-frontend VERSION=$$version; \
	mv backend/_build/eggplant ./eggplant
	@echo "Binary is available here: ./eggplant"

.PHONY: version
version:
	@hash=$$(git rev-parse HEAD 2>/dev/null); \
	if [ -z "$$hash" ]; then \
		echo "error-getting-version"; \
	else \
		git update-index -q --refresh 2>/dev/null; \
		if git diff --quiet && git diff --cached --quiet; then \
			echo "$$hash"; \
		else \
			echo "$$hash-dirty"; \
		fi; \
	fi
