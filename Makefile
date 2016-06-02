VERSION    := 0.1.0
BUILD_DEPS := github.com/tcnksm/ghr
DIST_OS    := linux darwin openbsd
DIST_ARCH  := amd64 386

all: build

include .lib.mk

EXECUTABLE                ?= $(REPO_NAME)
DOCKER_EXECUTABLE         := $(DOCKER_DIR)/$(EXECUTABLE)
LDFLAGS                   := '-X main.name=$(EXECUTABLE) -X main.version=$(VERSION)'
DIST_DIR                  := dist
DIST_TARGETS              := $(foreach os,$(DIST_OS),$(foreach arch,$(DIST_ARCH),$(DIST_DIR)/$(EXECUTABLE)_$(os)_$(arch)))
GO_BUILD                  := $(GO) build -v -ldflags $(LDFLAGS)
REPO_USER                 := joshuarubin
UI_DIR                    := ui
WEBPACK_TARGETS           := $(UI_DIR)/public/js/bundle.js $(UI_DIR)/public/css/style.css
APISERVER_GENERATE_TARGET := apiserver/bindata_assetfs.go
WEBPACK_CONFIG            := $(UI_DIR)/webpack.config.js
INDEX_FILE                := $(UI_DIR)/public/index.html
JSX_FILES                 := $(call find-files-with-extension,$(UI_DIR)/app/js,js)
SCSS_FILES                := $(wildcard $(UI_DIR)/app/sass/*.scss)
WEBPACK_DEPS              := $(WEBPACK_CONFIG) $(JSX_FILES) $(SCSS_FILES)

GHR ?= ghr

build: $(EXECUTABLE)

$(EXECUTABLE): .generate-stamp $(GO_FILES_NO_TESTS) .install-stamp
	$(GO_BUILD) -o $(EXECUTABLE) $(BASE_PKG)

$(DOCKER_EXECUTABLE): .generate-stamp $(GO_FILES_NO_TESTS) .install-stamp-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(DOCKER_EXECUTABLE) $(BASE_PKG)

image: .image-stamp

.image-stamp: $(DOCKER_EXECUTABLE) $(DOCKERFILE)
	$(DOCKER) build -t $(DOCKER_IMAGE) Docker
	@$(TOUCH) .image-stamp

clean::
	$(RM) -r $(EXECUTABLE) $(DOCKER_EXECUTABLE) .image-stamp $(DIST_DIR)

circle: test coveralls

fix-circle:: touch

touch:
	$(TOUCH) \
		.npm-install-stamp \
		$(WEBPACK_TARGETS) \
		.webpack-stamp \
		$(APISERVER_GENERATE_TARGET) \
		.generate-stamp

$(DIST_TARGETS): %: .generate-stamp $(GO_FILES_NO_TESTS) $(INSTALL_DEPS)
	CGO_ENABLED=0 GOOS=$(word 2,$(subst _, ,$@)) GOARCH=$(word 3,$(subst _, ,$@)) $(GO_BUILD) -o $@ $(BASE_PKG)

dist: $(DIST_TARGETS)

github-release: dist
	$(call check_defined,GITHUB_API_TOKEN)
	$(GHR) --token "$(GITHUB_API_TOKEN)" --username "$(REPO_USER)" --repository "$(REPO_NAME)" --replace "v$(VERSION)" $(DIST_DIR)

deploy:

webpack: .webpack-stamp

.webpack-stamp: $(WEBPACK_TARGETS)
	@$(TOUCH) .webpack-stamp

$(WEBPACK_TARGETS): .npm-install-stamp $(WEBPACK_DEPS)
	(cd $(UI_DIR) && npm run build)

$(APISERVER_GENERATE_TARGET): .webpack-stamp $(INDEX_FILE)
	(cd apiserver && go generate)

.generate-stamp: $(APISERVER_GENERATE_TARGET)
	@$(TOUCH) .generate-stamp

generate: .generate-stamp

npm-install: .npm-install-stamp

.npm-install-stamp: $(UI_DIR)/package.json
	(cd $(UI_DIR) && npm install)
	@$(TOUCH) .npm-install-stamp

watch:
	godo start --watch

.PHONY: all build image clean dist github-release deploy generate npm-install webpack watch touch
