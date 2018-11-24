PKG = github.com/k1LoW/tbls
COMMIT = $$(git describe --tags --always)
OSNAME=${shell uname -s}
ifeq ($(OSNAME),Darwin)
	DATE = $$(gdate --utc '+%Y-%m-%d_%H:%M:%S')
else
	DATE = $$(date --utc '+%Y-%m-%d_%H:%M:%S')
endif

GO ?= GO111MODULE=on go

BUILD_LDFLAGS = -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)
RELEASE_BUILD_LDFLAGS = -s -w $(BUILD_LDFLAGS)

default: test

ci: build test testdoc test_too_many_tables test_json

test:
	usql pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -f testdata/pg.sql
	usql my://root:mypass@localhost:33306/testdb -f testdata/my.sql
	usql my://root:mypass@localhost:33308/testdb -f testdata/my.sql
	sqlite3 $(PWD)/testdata/testdb.sqlite3 < testdata/sqlite.sql
	$(GO) test ./... -coverprofile=coverage.txt -covermode=count
	make testdoc

doc: build
	./tbls doc pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -a testdata/additional_data.yml -f sample/postgres
	./tbls doc my://root:mypass@localhost:33306/testdb -a testdata/additional_data.yml -f sample/mysql
	./tbls doc my://root:mypass@localhost:33308/testdb -a testdata/additional_data.yml -f sample/mysql8
	./tbls doc sq://$(PWD)/testdata/testdb.sqlite3 -a testdata/additional_data.yml -f sample/sqlite
	./tbls doc pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -a testdata/additional_data.yml -j -f sample/adjust
	./tbls doc my://root:mypass@localhost:33306/testdb -a testdata/additional_data.yml -t svg -f sample/svg

testdoc:
	./tbls diff pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -a testdata/additional_data.yml sample/postgres
	./tbls diff my://root:mypass@localhost:33306/testdb -a testdata/additional_data.yml sample/mysql
	./tbls diff my://root:mypass@localhost:33308/testdb -a testdata/additional_data.yml sample/mysql8
	./tbls diff sq://$(PWD)/testdata/testdb.sqlite3 -a testdata/additional_data.yml sample/sqlite
	./tbls diff pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -a testdata/additional_data.yml -j sample/adjust
	./tbls diff my://root:mypass@localhost:33306/testdb -a testdata/additional_data.yml -t svg sample/svg

test_too_many_tables:
	usql pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -f testdata/createdb_too_many.sql
	usql pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable -f testdata/createtable_too_many.sql
	ulimit -n 256 && ./tbls doc pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable -f /tmp
	ulimit -n 256 && ./tbls diff pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable /tmp

test_json:
	./tbls out my://root:mypass@localhost:33306/testdb -a testdata/additional_data.yml -t json > /tmp/tbls.json
	./tbls diff json:///tmp/tbls.json sample/mysql

test_env:
	env TBLS_DSN=my://root:mypass@localhost:33306/testdb TBLS_DOC_PATH=sample/mysql ./tbls doc -a testdata/additional_data.yml -f
	env TBLS_DSN=my://root:mypass@localhost:33306/testdb TBLS_DOC_PATH=sample/mysql ./tbls diff -a testdata/additional_data.yml

build:
	packr
	$(GO) build -ldflags="$(BUILD_LDFLAGS)"
	packr clean

depsdev:
	GO111MODULE=off go get golang.org/x/tools/cmd/cover
	GO111MODULE=off go get golang.org/x/lint/golint
	GO111MODULE=off go get github.com/motemen/gobump/cmd/gobump
	GO111MODULE=off go get github.com/Songmu/goxz/cmd/goxz
	GO111MODULE=off go get github.com/tcnksm/ghr
	GO111MODULE=off go get github.com/Songmu/ghch/cmd/ghch
	GO111MODULE=off go get github.com/xo/usql
	GO111MODULE=off go get github.com/gobuffalo/packr/packr

crossbuild: depsdev
	$(eval ver = v$(shell gobump show -r version/))
	packr
	GO111MODULE=on goxz -pv=$(ver) -arch=386,amd64 -build-ldflags="$(RELEASE_BUILD_LDFLAGS)" \
	  -d=./dist/$(ver)
	packr clean

prerelease:
	$(eval ver = v$(shell gobump show -r version/))
	GO111MODULE=on ghch -w -N $(ver)

release: crossbuild
	$(eval ver = v$(shell gobump show -r version/))
	GO111MODULE=on ghr -username k1LoW -replace $(ver) dist/$(ver)

.PHONY: default test
