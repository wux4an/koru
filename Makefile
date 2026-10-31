GO ?= go
PREFIX ?= $(HOME)/.local
.PHONY: build install test js-test native-test vet race clean release test-install
build:
	CGO_ENABLED=0 $(GO) build -trimpath -buildvcs=false -ldflags "-X main.commit=$$(git rev-parse --short=12 HEAD 2>/dev/null)" -o bin/koru ./cmd/koru
install: build
	install -d $(PREFIX)/bin
	install -m 755 bin/koru $(PREFIX)/bin/koru
test:
	$(GO) test ./...
js-test:
	node --test packages/javascript/core.test.cjs
native-test:
	gcc -std=c11 -Wall -Wextra -Werror -I runtime/native/include runtime/native/koru_plugin.c runtime/native/scope_test.c -o /tmp/koru-scope-test
	/tmp/koru-scope-test
vet:
	$(GO) vet ./...
race:
	$(GO) test -race ./...
clean:
	rm -rf bin artifacts dist
release:
	scripts/release.sh v$$(sed -n 's/^\tversion = "\(.*\)"$$/\1/p' cmd/koru/main.go) dist
test-install: release
	scripts/test-install.sh dist
