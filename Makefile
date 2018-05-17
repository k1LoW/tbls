PKG = github.com/k1LoW/tbls
COMMIT = $$(git describe --tags --always)
DATE = $$(date --utc '+%Y-%m-%d_%H:%M:%S')
BUILD_LDFLAGS = -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)
RELEASE_BUILD_LDFLAGS = -s -w $(BUILD_LDFLAGS)

default: test

test:
	usql pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -f test/pg.sql
	usql my://root:mypass@localhost:33306/testdb -f test/my.sql
	go test -cover -v $(shell go list ./... | grep -v vendor)

cover: depsdev
	goveralls -service=travis-ci

build:
	go build -ldflags="$(BUILD_LDFLAGS)"

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

depsdev:
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	go get github.com/golang/lint/golint
	go get github.com/motemen/gobump/cmd/gobump
	go get github.com/Songmu/goxz/cmd/goxz
	go get github.com/tcnksm/ghr
	go get github.com/Songmu/ghch/cmd/ghch
	go get -u github.com/xo/usql

crossbuild: deps depsdev
	$(eval ver = v$(shell gobump show -r version/))
	goxz -pv=$(ver) -os=linux -arch=386,amd64 -build-ldflags="$(RELEASE_BUILD_LDFLAGS)" \
	  -d=./dist/$(ver)

prerelease:
	$(eval ver = v$(shell gobump show -r version/))
	ghch -w -N ${ver}

release: crossbuild
	$(eval ver = v$(shell gobump show -r version/))
	ghr -username k1LoW -replace ${ver} dist/${ver}

.PHONY: default test deps cover
