YARN := corepack yarn

.PHONY: ci
ci: install lint type-check build

.PHONY: install
install:
	$(YARN) install --immutable

.PHONY: serve
serve: install
	$(YARN) serve --clearScreen false

.PHONY: serve-public
serve-public: install
	$(YARN) serve --host 0.0.0.0 --clearScreen false

.PHONY: lint
lint:
	$(YARN) lint

.PHONY: type-check
type-check:
	$(YARN) type-check

.PHONY: build
build:
	$(YARN) build
	echo "$(VERSION)" > .version
