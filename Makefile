ci: ci-backend ci-frontend

ci-backend:
	$(MAKE) -C backend ci

ci-frontend:
	$(MAKE) -C frontend ci

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

.PHONY: ci ci-backend ci-frontend version
