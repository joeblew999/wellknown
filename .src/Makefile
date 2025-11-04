.PHONY: help clone-all clone-pocketbase clone-gogen clone-presentator clone-jsonschema clone-goPocJson update-all update-pocketbase update-gogen update-presentator update-jsonschema update-goPocJson clean list status

## help: Show available make targets
help:
	@echo "Reference Code Management:"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/^## /  make /' | column -t -s ':'

.DEFAULT_GOAL := help

## clone-all: Clone all reference repositories at once
clone-all: clone-pocketbase clone-gogen clone-presentator clone-jsonschema clone-goPocJson

## clone-pocketbase: Clone PocketBase (official source for understanding record proxies)
clone-pocketbase:
	@echo "📦 Cloning PocketBase..."
	@if [ ! -d "pocketbase" ]; then \
		git clone https://github.com/pocketbase/pocketbase.git; \
		echo "✅ PocketBase cloned"; \
	else \
		echo "ℹ️  PocketBase already exists"; \
	fi

## clone-gogen: Clone pocketbase-gogen (type-safe code generator)
clone-gogen:
	@echo "📦 Cloning pocketbase-gogen..."
	@if [ ! -d "pocketbase-gogen" ]; then \
		git clone https://github.com/Snonky/pocketbase-gogen.git; \
		echo "✅ pocketbase-gogen cloned"; \
	else \
		echo "ℹ️  pocketbase-gogen already exists"; \
	fi

## clone-presentator: Clone Presentator (PocketBase library pattern example)
clone-presentator:
	@echo "📦 Cloning Presentator..."
	@if [ ! -d "presentator" ]; then \
		git clone https://github.com/presentator/presentator.git; \
		echo "✅ Presentator cloned"; \
	else \
		echo "ℹ️  Presentator already exists"; \
	fi

## clone-jsonschema: Clone jsonschema (validation library we use)
clone-jsonschema:
	@echo "📦 Cloning jsonschema..."
	@if [ ! -d "jsonschema" ]; then \
		git clone https://github.com/santhosh-tekuri/jsonschema.git; \
		echo "✅ jsonschema cloned"; \
	else \
		echo "ℹ️  jsonschema already exists"; \
	fi

## clone-goPocJson: Clone goPocJsonSchemaForm (dynamic form generation example)
clone-goPocJson:
	@echo "📦 Cloning goPocJsonSchemaForm..."
	@if [ ! -d "goPocJsonSchemaForm" ]; then \
		git clone https://github.com/warlockxins/goPocJsonSchemaForm.git; \
		echo "✅ goPocJsonSchemaForm cloned"; \
	else \
		echo "ℹ️  goPocJsonSchemaForm already exists"; \
	fi

## update-all: Update all repositories with git pull
update-all: update-pocketbase update-gogen update-presentator update-jsonschema update-goPocJson

## update-pocketbase: Update PocketBase repository
update-pocketbase:
	@if [ -d "pocketbase" ]; then \
		echo "⬇️  Updating PocketBase..."; \
		cd pocketbase && git pull; \
		echo "✅ PocketBase updated"; \
	else \
		echo "❌ PocketBase not found. Run: make clone-pocketbase"; \
	fi

## update-gogen: Update pocketbase-gogen repository
update-gogen:
	@if [ -d "pocketbase-gogen" ]; then \
		echo "⬇️  Updating pocketbase-gogen..."; \
		cd pocketbase-gogen && git pull; \
		echo "✅ pocketbase-gogen updated"; \
	else \
		echo "❌ pocketbase-gogen not found. Run: make clone-gogen"; \
	fi

## update-presentator: Update Presentator repository
update-presentator:
	@if [ -d "presentator" ]; then \
		echo "⬇️  Updating Presentator..."; \
		cd presentator && git pull; \
		echo "✅ Presentator updated"; \
	else \
		echo "❌ Presentator not found. Run: make clone-presentator"; \
	fi

## update-jsonschema: Update jsonschema repository
update-jsonschema:
	@if [ -d "jsonschema" ]; then \
		echo "⬇️  Updating jsonschema..."; \
		cd jsonschema && git pull; \
		echo "✅ jsonschema updated"; \
	else \
		echo "❌ jsonschema not found. Run: make clone-jsonschema"; \
	fi

## update-goPocJson: Update goPocJsonSchemaForm repository
update-goPocJson:
	@if [ -d "goPocJsonSchemaForm" ]; then \
		echo "⬇️  Updating goPocJsonSchemaForm..."; \
		cd goPocJsonSchemaForm && git pull; \
		echo "✅ goPocJsonSchemaForm updated"; \
	else \
		echo "❌ goPocJsonSchemaForm not found. Run: make clone-goPocJson"; \
	fi

## clean: Remove all cloned repositories (WARNING: destructive!)
clean:
	@echo "⚠️  This will delete all cloned reference repositories"
	@read -p "Are you sure? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		rm -rf pocketbase pocketbase-gogen presentator jsonschema goPocJsonSchemaForm; \
		echo "✅ All repositories removed"; \
	else \
		echo "❌ Cancelled"; \
	fi

## list: List which repositories are currently cloned
list:
	@echo "📚 Reference Repositories:"
	@echo ""
	@for dir in pocketbase pocketbase-gogen presentator jsonschema goPocJsonSchemaForm; do \
		if [ -d "$$dir" ]; then \
			echo "  ✅ $$dir"; \
		else \
			echo "  ❌ $$dir (not cloned)"; \
		fi; \
	done

## status: Show git status and latest commit for all cloned repositories
status:
	@echo "📊 Repository Status:"
	@echo ""
	@for dir in pocketbase pocketbase-gogen presentator jsonschema goPocJsonSchemaForm; do \
		if [ -d "$$dir" ]; then \
			echo "=== $$dir ==="; \
			cd $$dir && git status --short && git log --oneline -1 && cd ..; \
			echo ""; \
		fi; \
	done
