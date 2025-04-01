# SPDX-FileCopyrightText: (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

# common.mk - common targets for library repo

# Makefile Style Guide:
# - Help will be generated from ## comments at end of any target line
# - Use smooth parens $() for variables over curly brackets ${} for consistency
# - Continuation lines (after an \ on previous line) should start with spaces
#   not tabs - this will cause editor highligting to point out editing mistakes
# - When creating targets that run a lint or similar testing tool, print the
#   tool version first so that issues with versions in CI or other remote
#   environments can be caught

# Optionally include tool version checks, not used in Docker builds
ifeq ($(TOOL_VERSION_CHECK), 1)
	include ../version.mk
endif

#### Variables ####

# Shell config variable
SHELL	:= bash -eu -o pipefail

# Path variables
OUT_DIR	:= out

#### Path Target ####

$(OUT_DIR): ## Create out directory
	mkdir -p $(OUT_DIR)

#### Python venv Target ####
VENV_NAME	:= venv-env

$(VENV_NAME): requirements.txt ## Create Python venv
	python3 -m venv $@ ;\
  set +u; . ./$@/bin/activate; set -u ;\
  python -m pip install --upgrade pip ;\
  python -m pip install -r requirements.txt

## Generic linter targets
# https://pypi.org/project/reuse/
license: $(VENV_NAME) ## Check licensing with the reuse tool
	set +u; . ./$</bin/activate; set -u ;\
  reuse --version ;\
  reuse --root . lint


yamllint: $(VENV_NAME) ## Lint YAML files
	. ./$</bin/activate; set -u ;\
  yamllint --version ;\
  yamllint -d '{extends: default, rules: {line-length: {max: 99}}, ignore: [$(YAML_IGNORE)]}' -s $(YAML_FILES)

mdlint: ## Link MD files
	markdownlint --version ;\
	markdownlint "**/*.md" -c ../.markdownlint.yml

#### Clean Targets ###
common-clean: ## Delete build and vendor directories
	rm -rf $(OUT_DIR) vendor

clean-venv: ## Delete Python venv
	rm -rf "$(VENV_NAME)"

clean-all: common-clean clean-venv ## Delete all built artifacts and downloaded tools

#### Help Target ####

help: ## Print help for each target
	@echo $(PROJECT_NAME) make targets
	@echo "Target               Makefile:Line    Description"
	@echo "-------------------- ---------------- -----------------------------------------"
	@grep -H -n '^[[:alnum:]_-]*:.* ##' $(MAKEFILE_LIST) \
    | sort -t ":" -k 3 \
    | awk 'BEGIN  {FS=":"}; {sub(".* ## ", "", $$4)}; {printf "%-20s %-16s %s\n", $$3, $$1 ":" $$2, $$4};'