version:
	@git update-index -q --refresh 2>/dev/null; printf '%s%s\n' "$$(git rev-parse HEAD)" "$$(git diff --quiet && git diff --cached --quiet || echo -dirty)"

.PHONY: version
