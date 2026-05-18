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

.PHONY: version
