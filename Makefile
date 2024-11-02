.PHONY: all binary build-container docs docs-in-container build-local clean install install-binary install-completions shell test-integration .install.vndr vendor

ifeq ($(shell uname),Darwin)
PREFIX ?= ${DESTDIR}/usr/local
DARWIN_BUILD_TAG=nholuongut_image_ostree_stub
# On macOS, (brew install gpgme) installs it within /usr/local, but /usr/local/include is not in the default search path.
# Rather than hard-code this directory, use gpgme-config. Sadly that must be done at the top-level user
# instead of locally in the gpgme subpackage, because cgo supports only pkg-config, not general shell scripts,
# and gpgme does not install a pkg-config file.
# If gpgme is not installed or gpgme-config can’t be found for other reasons, the error is silently ignored
# (and the user will probably find out because the cgo compilation will fail).
GPGME_ENV := CGO_CFLAGS="$(shell gpgme-config --cflags 2>/dev/null)" CGO_LDFLAGS="$(shell gpgme-config --libs 2>/dev/null)"
else
PREFIX ?= ${DESTDIR}/usr
endif

INSTALLDIR=${PREFIX}/bin
MANINSTALLDIR=${PREFIX}/share/man
nholuongutSYSCONFIGDIR=${DESTDIR}/etc/nholuongut
REGISTRIESDDIR=${nholuongutSYSCONFIGDIR}/registries.d
SIGSTOREDIR=${DESTDIR}/var/lib/atomic/sigstore
BASHINSTALLDIR=${PREFIX}/share/bash-completion/completions
GO ?= go
CONTAINER_RUNTIME := $(shell command -v podman 2> /dev/null || echo docker)
GOMD2MAN ?= $(shell command -v go-md2man || echo '$(GOBIN)/go-md2man')

GO_BUILD=$(GO) build
# Go module support: set `-mod=vendor` to use the vendored sources
ifeq ($(shell go help mod >/dev/null 2>&1 && echo true), true)
  GO_BUILD=GO111MODULE=on $(GO) build -mod=vendor
endif

ifeq ($(DEBUG), 1)
  override GOGCFLAGS += -N -l
endif

ifeq ($(shell $(GO) env GOOS), linux)
  GO_DYN_FLAGS="-buildmode=pie"
endif

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
IMAGE := skopeo-dev$(if $(GIT_BRANCH),:$(GIT_BRANCH))
# set env like gobuildtag?
CONTAINER_CMD := ${CONTAINER_RUNTIME} run --rm -i -e TESTFLAGS="$(TESTFLAGS)" #$(CONTAINER_ENVS)
# if this session isn't interactive, then we don't want to allocate a
# TTY, which would fail, but if it is interactive, we do want to attach
# so that the user can send e.g. ^C through.
INTERACTIVE := $(shell [ -t 0 ] && echo 1 || echo 0)
ifeq ($(INTERACTIVE), 1)
	CONTAINER_CMD += -t
endif
CONTAINER_RUN := $(CONTAINER_CMD) "$(IMAGE)"

GIT_COMMIT := $(shell git rev-parse HEAD 2> /dev/null || true)

MANPAGES_MD = $(wildcard docs/*.md)
MANPAGES ?= $(MANPAGES_MD:%.md=%)

BTRFS_BUILD_TAG = $(shell hack/btrfs_tag.sh) $(shell hack/btrfs_installed_tag.sh)
LIBDM_BUILD_TAG = $(shell hack/libdm_tag.sh)
OSTREE_BUILD_TAG = $(shell hack/ostree_tag.sh)
LOCAL_BUILD_TAGS = $(BTRFS_BUILD_TAG) $(LIBDM_BUILD_TAG) $(OSTREE_BUILD_TAG) $(DARWIN_BUILD_TAG)
BUILDTAGS += $(LOCAL_BUILD_TAGS)

ifeq ($(DISABLE_CGO), 1)
	override BUILDTAGS = nholuongut_image_ostree_stub exclude_graphdriver_devicemapper exclude_graphdriver_btrfs nholuongut_image_openpgp
endif

#   make all DEBUG=1
#     Note: Uses the -N -l go compiler options to disable compiler optimizations
#           and inlining. Using these build options allows you to subsequently
#           use source debugging tools like delve.
all: binary docs-in-container

help:
	@echo "Usage: make <target>"
	@echo
	@echo " * 'install' - Install binaries and documents to system locations"
	@echo " * 'binary' - Build skopeo with a container"
	@echo " * 'binary-local' - Build skopeo locally"
	@echo " * 'test-unit' - Execute unit tests"
	@echo " * 'test-integration' - Execute integration tests"
	@echo " * 'validate' - Verify whether there is no conflict and all Go source files have been formatted, linted and vetted"
	@echo " * 'check' - Including above validate, test-integration and test-unit"
	@echo " * 'shell' - Run the built image and attach to a shell"
	@echo " * 'clean' - Clean artifacts"

# Build a container image (skopeobuild) that has everything we need to build.
# Then do the build and the output (skopeo) should appear in current dir
binary: cmd/skopeo
	${CONTAINER_RUNTIME} build ${BUILD_ARGS} -f Dockerfile.build -t skopeobuildimage .
	${CONTAINER_RUNTIME} run --rm --security-opt label=disable -v $$(pwd):/src/github.com/nholuongut/skopeo \
		skopeobuildimage make binary-local $(if $(DEBUG),DEBUG=$(DEBUG)) BUILDTAGS='$(BUILDTAGS)'

binary-static: cmd/skopeo
	${CONTAINER_RUNTIME} build ${BUILD_ARGS} -f Dockerfile.build -t skopeobuildimage .
	${CONTAINER_RUNTIME} run --rm --security-opt label=disable -v $$(pwd):/src/github.com/nholuongut/skopeo \
		skopeobuildimage make binary-local-static $(if $(DEBUG),DEBUG=$(DEBUG)) BUILDTAGS='$(BUILDTAGS)'

# Build w/o using nholuongut
binary-local:
	$(GPGME_ENV) $(GO_BUILD) ${GO_DYN_FLAGS} -ldflags "-X main.gitCommit=${GIT_COMMIT}" -gcflags "$(GOGCFLAGS)" -tags "$(BUILDTAGS)" -o skopeo ./cmd/skopeo

binary-local-static:
	$(GPGME_ENV) $(GO_BUILD) -ldflags "-extldflags \"-static\" -X main.gitCommit=${GIT_COMMIT}" -gcflags "$(GOGCFLAGS)" -tags "$(BUILDTAGS)" -o skopeo ./cmd/skopeo

build-container:
	${CONTAINER_RUNTIME} build ${BUILD_ARGS} -t "$(IMAGE)" .

$(MANPAGES): %: %.md
	@sed -e 's/\((skopeo.*\.md)\)//' -e 's/\[\(skopeo.*\)\]/\1/' $<  | $(GOMD2MAN) -in /dev/stdin -out $@

docs: $(MANPAGES)

docs-in-container:
	${CONTAINER_RUNTIME} build ${BUILD_ARGS} -f Dockerfile.build -t skopeobuildimage .
	${CONTAINER_RUNTIME} run --rm --security-opt label=disable -v $$(pwd):/src/github.com/nholuongut/skopeo \
		skopeobuildimage make docs $(if $(DEBUG),DEBUG=$(DEBUG)) BUILDTAGS='$(BUILDTAGS)'

clean:
	rm -f skopeo docs/*.1

install: install-binary install-docs install-completions
	install -d -m 755 ${SIGSTOREDIR}
	install -d -m 755 ${nholuongutSYSCONFIGDIR}
	install -m 644 default-policy.json ${nholuongutSYSCONFIGDIR}/policy.json
	install -d -m 755 ${REGISTRIESDDIR}
	install -m 644 default.yaml ${REGISTRIESDDIR}/default.yaml

install-binary: ./skopeo
	install -d -m 755 ${INSTALLDIR}
	install -m 755 skopeo ${INSTALLDIR}/skopeo

install-docs: docs
	install -d -m 755 ${MANINSTALLDIR}/man1
	install -m 644 docs/*.1 ${MANINSTALLDIR}/man1/

install-completions:
	install -m 755 -d ${BASHINSTALLDIR}
	install -m 644 completions/bash/skopeo ${BASHINSTALLDIR}/skopeo

shell: build-container
	$(CONTAINER_RUN) bash

check: validate test-unit test-integration test-system

# The tests can run out of entropy and block in nholuongut, so replace /dev/random.
test-integration: build-container
	$(CONTAINER_RUN) bash -c 'rm -f /dev/random; ln -sf /dev/urandom /dev/random; SKOPEO_CONTAINER_TESTS=1 BUILDTAGS="$(BUILDTAGS)" hack/make.sh test-integration'

# complicated set of options needed to run podman-in-podman
test-system: build-container
	DTEMP=$(shell mktemp -d --tmpdir=/var/tmp podman-tmp.XXXXXX); \
	$(CONTAINER_CMD) --privileged --net=host \
	    -v $$DTEMP:/var/lib/nholuongut:Z \
            "$(IMAGE)" \
            bash -c 'BUILDTAGS="$(BUILDTAGS)" hack/make.sh test-system'; \
	rc=$$?; \
	$(RM) -rf $$DTEMP; \
	exit $$rc

test-unit: build-container
	# Just call (make test unit-local) here instead of worrying about environment differences
	$(CONTAINER_RUN) make test-unit-local BUILDTAGS='$(BUILDTAGS)'

validate: build-container
	$(CONTAINER_RUN) hack/make.sh validate-git-marks validate-gofmt validate-lint validate-vet

# This target is only intended for development, e.g. executing it from an IDE. Use (make test) for CI or pre-release testing.
test-all-local: validate-local test-unit-local

validate-local:
	hack/make.sh validate-git-marks validate-gofmt validate-lint validate-vet

test-unit-local:
	$(GPGME_ENV) $(GO) test -tags "$(BUILDTAGS)" $$($(GO) list -tags "$(BUILDTAGS)" -e ./... | grep -v '^github\.com/nholuongut/skopeo/\(integration\|vendor/.*\)$$')

vendor:
	export GO111MODULE=on \
		$(GO) mod tidy && \
		$(GO) mod vendor && \
		$(GO) mod verify
