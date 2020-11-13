# `grep -v` does not work on travis. No time to find out why -- xp 2019 Feb 22
PKGS := $(shell go list ./... | grep -v "^github.com/openacid/polyarray/\(vendor\|prototype\)")

# PKGS := github.com/openacid/polyarray/array \
#         github.com/openacid/polyarray/bit \
#         github.com/openacid/polyarray/trie \

SRCDIRS := $(shell go list -f '{{.Dir}}' $(PKGS))

# gofmt check vendor dir. we need to skip vendor manually
GOFILES := $(shell find $(SRCDIRS) -not -path "*/vendor/*" -name "*.go")
GO := go

check: test vet gofmt misspell unconvert staticcheck ineffassign unparam

travis: test vet gofmt misspell unconvert ineffassign unparam

test:
	$(GO) test -v -tags debug $(PKGS)
	$(GO) test -v             $(PKGS)

vet: | test
	$(GO) vet $(PKGS)

staticcheck:
	$(GO) get honnef.co/go/tools/cmd/staticcheck
	# ST1016: methods on the same type should have the same receiver name
	#         .pb.go have this issue.
	staticcheck -checks all,-ST1016 $(PKGS)

misspell:
	$(GO) get github.com/client9/misspell/cmd/misspell
	find $(SRCDIRS) -name '*.go' -or -name '*.md' | grep -v "\bvendor/" | xargs misspell \
		-locale US \
		-error
	misspell \
		-locale US \
		-error \
		*.md *.go

unconvert:
	$(GO) get github.com/mdempsky/unconvert
	unconvert -v $(PKGS)

ineffassign:
	$(GO) get github.com/gordonklaus/ineffassign
	find $(SRCDIRS) -name '*.go' | grep -v "\bvendor/" | xargs ineffassign

pedantic: check errcheck

unparam:
	$(GO) get mvdan.cc/unparam
	unparam ./...

errcheck:
	$(GO) get github.com/kisielk/errcheck
	errcheck $(PKGS)

gofmt:
	@echo Checking code is gofmted
	@test -z "$(shell gofmt -s -l -d -e $(GOFILES) | tee /dev/stderr)"

ben: test
	$(GO) test ./... -run=none -bench=. -benchmem

gen:
	$(GO) generate ./...

doc:
	# $(GO) get github.com/robertkrimen/godocdown/godocdown
	godocdown . > docs/polyarray.md
	cat docs/polyarray.md | awk '/^package /,/^## Usage/' | grep -v '^## Usage' > docs/polyarray-package.md


readme: doc
	python ./scripts/build_readme.py
	# brew install nodejs
	# npm install -g doctoc
	doctoc --title '' --github README.md

fix:
	gofmt -s -w $(GOFILES)
	unconvert -v -apply $(PKGS)


# local coverage
coverage:
	$(GO) test -covermode=count -coverprofile=coverage.out $(PKGS)
	go tool cover -func=coverage.out
	# go tool cover -html=coverage.out

# send coverage to coveralls
coveralls:
	$(GO) get golang.org/x/tools/cmd/cover
	$(GO) get github.com/mattn/goveralls
	$(GO) test -covermode=count -coverprofile=coverage.out $(PKGS)
	goveralls -ignore='polyarray.pb.go' -coverprofile=coverage.out -service=travis-ci
	# goveralls -coverprofile=coverage.out -service=travis-ci -repotoken $$COVERALLS_TOKEN
