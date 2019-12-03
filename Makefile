app_name	:= connector
app_ver		?= 0.0.1
app_cmd     := ./cmd/connector/

VERSION_PATH	:= github.com/leitstand/leitstand-powerdns/pkg/version
VERSION_PATH	:= $(shell echo "$(VERSION_PATH)" | head -n 1 | awk '{printf("%s", $$1);}')
LDFLAGS		:= -X "$(VERSION_PATH).VERSION=$(app_ver)" $(LDFLAGS)

# Just in case there is an extra space at the end of the line.
app_name	:= $(shell echo "$(app_name)" | head -n 1 | awk '{printf("%s", $$1);}')
app_ver		:= $(shell echo "$(app_ver)" | head -n 1 | awk '{printf("%s", $$1);}')

OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))

all: gocmds

linters:
	@echo "[run] linters"
	@# We do this instead of a simple `go fmt ...` because (at least in the
	@# begining) it's better too see the changes than blindly run it.
	@echo "gofmt -d -e ./cmd/ ./pkg/"; \
		fmt_out=`gofmt -d -e ./cmd/ ./pkg/` || exit 1; \
		[ -z "$$fmt_out" ] || { \
			echo "$$fmt_out"; \
			echo "#"; \
			echo "# If you want a quick fix just run: go fmt ."; \
			echo "#"; \
			exit 1; \
		};
	@which golint > /dev/null || { \
		echo "#"; \
		echo "# Either you don't have golint installed or it's not accessible."; \
		echo "#"; \
		echo "# Make sure you have \$$GOPATH set up correctly and that \$$GOPATH/bin is included in your \$$PATH,"; \
		echo "# see https://golang.org/doc/code.html#GOPATH & https://github.com/golang/go/wiki/GOPATH ."; \
		echo "#"; \
		echo "# After that run: go get -u golang.org/x/lint/golint"; \
		echo "# see https://github.com/golang/lint ."; \
		echo "#"; \
		exit 1; \
	};
	golint -set_exit_status ./cmd/... ./pkg/...
	go vet -mod=vendor ./cmd/... ./pkg/...

test:
	@echo "[run] tests"
	@mkdir -p $(BUILD_DIR)
	go test -mod=vendor -coverprofile=./bin/cover.out -cover ./...

build-%:
	@echo	"[run] build-OS_ARCH"
	@$(MAKE) build                        \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

BUILD_DIR := bin/$(OS)_$(ARCH)

build:
	@echo build $(OS)_$(ARCH) $(app_ver)
	@mkdir -p $(BUILD_DIR)
	env GOOS=$(OS) GOARCH=$(ARCH) go build -mod=vendor -o $(BUILD_DIR)/$(app_name) -ldflags '$(LDFLAGS)' $(app_cmd)

gocmds: linters test build-darwin_amd64 build-linux_amd64


.PHONY: clean
clean:
	@echo "[run] clean"
	@- rm -rf bin

build-swagger:
	@echo generate swagger
	$(GOPATH)/bin/swag init -g $(app_cmd)$(app_name).go -o ./doc
	rm ./doc/docs.go

docker-build: build-linux_amd64
	docker build -t leitstand/powerdns-connector:latest .

docker-run: docker-build
	docker run --rm --name leitstand-powerdns-connector -p 19991:19991 leitstand/powerdns-connector