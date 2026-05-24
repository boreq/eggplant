YARN := corepack yarn

ci: install lint type-check build

install:
	$(YARN) install --immutable

lint:
	$(YARN) lint

type-check:
	$(YARN) type-check

build:
	$(YARN) build

.PHONY: ci install lint type-check build
