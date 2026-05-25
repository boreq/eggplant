YARN := corepack yarn

.PHONY: ci
ci: install lint type-check build

.PHONY: install
install:
	$(YARN) install --immutable

.PHONY: lint
lint:
	$(YARN) lint

.PHONY: type-check
type-check:
	$(YARN) type-check

.PHONY: build
build:
	$(YARN) build
