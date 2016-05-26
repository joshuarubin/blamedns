SHELL := $(SHELL) -e

TOUCH     ?= touch
GO        ?= go
FIND      ?= find
TRUE      ?= true
RM        ?= rm -f
MV        ?= mv
DOCKER    ?= docker
GODEP     ?= godep

METALINT           := gometalinter --cyclo-over=10 --deadline=10s -t
REPO_NAME          := $(notdir $(CURDIR))
SRC_DIR            := $(realpath $(CURDIR)/../..)
BASE_PKG           := jrubin.io/$(REPO_NAME)
EXCLUDE_DIRS       := ./vendor ./Godeps ./.git
FIRST_GOPATH       := $(firstword $(subst :, ,$(GOPATH)))

GOTEST             ?= $(GO) test # use "gt" if you want results cached for quicker testing
GO_LINTERS         ?= golint go\_vet # \_ is replaced with a space in the command
DOCKER_DIR         ?= Docker
DOCKERFILE         ?= $(DOCKER_DIR)/Dockerfile
DOCKER_IMAGE       ?= joshuarubin/$(REPO_NAME)
DIST_OS            ?= linux
DIST_ARCH          ?= amd64

BUILD_DEPS += rsc.io/gt github.com/tools/godep

INSTALL_DEPS        := $(foreach os,$(DIST_OS),$(foreach arch,$(DIST_ARCH),.install-stamp-$(os)-$(arch)))

export GO15VENDOREXPERIMENT=1

# return test_string only if it does not contain any of the exclude_strings
#
# $(call exclude,test_string,exclude_strings
exclude = $(foreach 1,$1,$(if $(sort $(foreach 2,$2,$(if $(findstring $2,$1),$1))),,$1))

# GO_PKGS - a list of go packages excluding vendor
GO_PKGS := \
	$(strip \
	$(foreach pkg, \
		$(shell $(GO) list ./... 2>/dev/null), \
		$(call exclude,$(pkg),/vendor/)))

# ALL_DIRS - a list of directories, recursive, from $(CURDIR) excluding
# anything containing any of $(EXCLUDE_DIRS)
ALL_DIRS := \
	$(patsubst ./%, %, \
	$(sort \
	$(foreach dir, \
		$(shell $(FIND) . -type d), \
		$(call exclude,$(dir),$(EXCLUDE_DIRS)))))

# get all files in $(ALL_DIRS) that have the given extension
#
# $(call find-files-with-extension,$(extension))
find-files-with-extension = \
	$(strip \
	$(patsubst ./%, %, \
	$(foreach dir, \
		$(ALL_DIRS), \
		$(wildcard $(dir)/*.$1))))

# GO_FILES - a list of all .go files in any directory in $(ALL_DIRS)
GO_FILES := $(call find-files-with-extension,go)

# GO_FILES_NO_TESTS - $(GO_FILES) with test files removed
GO_FILES_NO_TESTS := $(filter-out %_test.go,$(GO_FILES))

define \n


endef

check_defined  = $(foreach 1,$1,$(if $(value $1),, $(error $1 is undefined)))
file_exists    = $(if $(wildcard $1),$1)

lint:
	$(foreach pkg, $(GO_PKGS), $(foreach cmd, $(GO_LINTERS), $(subst \_, , $(cmd)) $(pkg) || $(TRUE)${\n}))

metalint:
	$(foreach pkg, $(GO_PKGS), $(METALINT) $(SRC_DIR)/$(pkg) || $(TRUE)${\n})

test: $(GO_FILES)
	$(GOTEST) -race $(GO_PKGS)

save:
	$(RM) -r ./Godeps ./vendor
	GOOS=linux GOARCH=amd64 $(GODEP) save ./...

clean::
	$(RM) ./.install-stamp $(INSTALL_DEPS)

install:: .install-stamp $(INSTALL_DEPS)

.install-stamp: $(GO_FILES)
	$(GO) install -v ./... || $(TRUE)
	@$(TOUCH) .install-stamp

$(INSTALL_DEPS): %: $(GO_FILES)
	CGO_ENABLED=0 GOOS=$(word 3,$(subst -, ,$@)) GOARCH=$(word 4,$(subst -, ,$@)) $(GO) install -v ./... || $(TRUE)
	@$(TOUCH) $@

.push_image: .image-stamp
	$(call check_defined,VERSION)

	$(DOCKER) push $(DOCKER_IMAGE):latest
	$(DOCKER) tag $(DOCKER_IMAGE):latest $(DOCKER_IMAGE):v$(VERSION)
	$(DOCKER) push $(DOCKER_IMAGE):v$(VERSION)

	@$(TOUCH) .push_image

go_bin_exists        = $(call file_exists,$(FIRST_GOPATH)/bin/$(notdir $1))
go_bin_not_installed = $(if $(call go_bin_exists,$1),,$1)
go_deps_to_install  := $(strip $(foreach dep,$(BUILD_DEPS),$(call go_bin_not_installed,$(dep))))

build_deps:
	$(if $(go_deps_to_install),$(GO) get -v $(go_deps_to_install))

version:
	$(call check_defined,VERSION)
	@$(ECHO) $(VERSION)

.PHONY: lint metalint test save clean install build_deps version
