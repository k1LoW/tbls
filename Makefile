PKG = github.com/k1LoW/tbls
COMMIT = $$(git describe --tags --always)
OSNAME=${shell uname -s}
ifeq ($(OSNAME),Darwin)
	DATE = $$(gdate --utc '+%Y-%m-%d_%H:%M:%S')
else
	DATE = $$(date --utc '+%Y-%m-%d_%H:%M:%S')
endif

BUILD_LDFLAGS = -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)
RELEASE_BUILD_LDFLAGS = -s -w $(BUILD_LDFLAGS)

default: test

test:
	usql pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -f test/pg.sql
	usql my://root:mypass@localhost:33306/testdb -f test/my.sql
	usql my://root:mypass@localhost:33308/testdb -f test/my.sql
	go test -cover -v $(shell go list ./... | grep -v vendor)
	make testdoc

cover: depsdev
	goveralls -service=travis-ci

template:
	go generate

doc: build
	./tbls doc pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -a test/additional_data.yml -f sample/postgres
	./tbls doc my://root:mypass@localhost:33306/testdb -a test/additional_data.yml -f sample/mysql
	./tbls doc my://root:mypass@localhost:33308/testdb -a test/additional_data.yml -f sample/mysql8

testdoc: build
	$(eval DIFF := $(shell ./tbls diff pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -a test/additional_data.yml sample/postgres))
	@test -z "$(DIFF)" || (echo "document does not match database." && ./tbls diff pg://postgres:pgpass@localhost:55432?sslmode=disable -a test/additional_data.yml sample/postgres && exit 1)
	$(eval DIFF := $(shell ./tbls diff my://root:mypass@localhost:33306/testdb -a test/additional_data.yml sample/mysql))
	@test -z "$(DIFF)" || (echo "document does not match database." && ./tbls diff my://root:mypass@localhost:33306/testdb -a test/additional_data.yml sample/mysql && exit 1)
	$(eval DIFF := $(shell ./tbls diff my://root:mypass@localhost:33308/testdb -a test/additional_data.yml sample/mysql8))
	@test -z "$(DIFF)" || (echo "document does not match database." && ./tbls diff my://root:mypass@localhost:33308/testdb -a test/additional_data.yml sample/mysql8 && exit 1)

build: template
	go build -ldflags="$(BUILD_LDFLAGS)"

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

depsdev: deps
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	go get github.com/golang/lint/golint
	go get github.com/motemen/gobump/cmd/gobump
	go get github.com/Songmu/goxz/cmd/goxz
	go get github.com/tcnksm/ghr
	go get github.com/Songmu/ghch/cmd/ghch
	go get github.com/xo/usql
	go get github.com/jessevdk/go-assets-builder

crossbuild: deps depsdev
	$(eval ver = v$(shell gobump show -r version/))
	goxz -pv=$(ver) -arch=386,amd64 -build-ldflags="$(RELEASE_BUILD_LDFLAGS)" \
	  -d=./dist/$(ver)

prerelease:
	$(eval ver = v$(shell gobump show -r version/))
	ghch -w -N ${ver}

release: crossbuild
	$(eval ver = v$(shell gobump show -r version/))
	ghr -username k1LoW -replace ${ver} dist/${ver}

.PHONY: default test deps cover
