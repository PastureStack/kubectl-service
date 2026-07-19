TARGETS := $(shell ls scripts 2>/dev/null)

.PHONY: $(TARGETS) deps trash trash-keep

$(TARGETS):
	@echo "Vendored dependency; standalone dapper execution is disabled in this maintenance fork."

trash:
	@echo "Dependencies are vendored; no legacy trash dependency sync is required."

trash-keep: trash

deps: trash

.DEFAULT_GOAL := ci
