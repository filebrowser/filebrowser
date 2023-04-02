SHELL := /usr/bin/env bash
DATE ?= $(shell date +%FT%T%z)
BASE_PATH := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION ?= $(shell git describe --tags --always --match=v* 2> /dev/null || \
           			cat $(CURDIR)/.version 2> /dev/null || echo v0)
VERSION_HASH = $(shell git rev-parse HEAD)
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

go = GOGC=off go
MODULE = $(shell env GO111MODULE=on go list -m)

# printing
# $Q (quiet) is used in the targets as a replacer for @.
# This macro helps to print the command for debugging by setting V to 1. Example `make test-unit V=1`
V = 0
Q = $(if $(filter 1,$V),,@)
# $M is a macro to print a colored ▶ character. Example `$(info $(M) running coverage tests…)` will print "▶ running coverage tests…"
M = $(shell printf "\033[34;1m▶\033[0m")

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

define global_option
    printf "  ${YELLOW}%-20s${GREEN}%s${RESET}\n" $(1) $(2)
endef
