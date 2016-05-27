VERSION    := 0.1.0
BUILD_DEPS := github.com/tcnksm/ghr
DIST_OS    := linux darwin openbsd
DIST_ARCH  := amd64 386

all: build

include .lib.mk

EXECUTABLE        ?= $(REPO_NAME)
DOCKER_EXECUTABLE := $(DOCKER_DIR)/$(EXECUTABLE)
LDFLAGS           := '-X main.name=$(EXECUTABLE) -X main.version=$(VERSION)'
DIST_DIR          := dist
DIST_TARGETS      := $(foreach os,$(DIST_OS),$(foreach arch,$(DIST_ARCH),$(DIST_DIR)/$(EXECUTABLE)_$(os)_$(arch)))
GO_BUILD          := $(GO) build -v -ldflags $(LDFLAGS)
REPO_USER         := joshuarubin

GHR ?= ghr

build: $(EXECUTABLE)

$(EXECUTABLE): $(GO_FILES_NO_TESTS) .install-stamp
	$(GO_BUILD) -o $(EXECUTABLE) $(BASE_PKG)

$(DOCKER_EXECUTABLE): $(GO_FILES_NO_TESTS) .install-stamp-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(DOCKER_EXECUTABLE) $(BASE_PKG)

image: .image-stamp

.image-stamp: $(DOCKER_EXECUTABLE) $(DOCKERFILE)
	$(DOCKER) build -t $(DOCKER_IMAGE) Docker
	@touch .image-stamp

clean::
	$(RM) -r $(EXECUTABLE) $(DOCKER_EXECUTABLE) .image-stamp $(DIST_DIR)

circle: coveralls

$(DIST_TARGETS): %: $(GO_FILES_NO_TESTS) $(INSTALL_DEPS)
	CGO_ENABLED=0 GOOS=$(word 2,$(subst _, ,$@)) GOARCH=$(word 3,$(subst _, ,$@)) $(GO_BUILD) -o $@ $(BASE_PKG)

dist: $(DIST_TARGETS)

github-release: dist
	$(call check_defined,GITHUB_API_TOKEN)
	$(GHR) --token "$(GITHUB_API_TOKEN)" --username "$(REPO_USER)" --repository "$(REPO_NAME)" --replace "v$(VERSION)" $(DIST_DIR)

deploy:

.PHONY: all build image clean dist github-release deploy
