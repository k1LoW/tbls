PKG = github.com/k1LoW/tbls
COMMIT = $$(git describe --tags --always)
OSNAME=${shell uname -s}
ifeq ($(OSNAME),Darwin)
	SED = gsed
	DATE = $$(gdate --utc '+%Y-%m-%d_%H:%M:%S')
else
	SED = sed
	DATE = $$(date --utc '+%Y-%m-%d_%H:%M:%S')
endif

export GO111MODULE=on
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy
export AWS_DEFAULT_REGION=ap-northeast-1

BUILD_LDFLAGS = -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)

default: test

ci: depsdev build db test testdoc test_too_many_tables test_json sec doc

db:
	usql pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -f testdata/pg.sql
	usql my://root:mypass@localhost:33306/testdb -f testdata/my.sql
	usql my://root:mypass@localhost:33308/testdb -f testdata/my.sql
	sqlite3 $(PWD)/testdata/testdb.sqlite3 < testdata/sqlite.sql
	usql ms://SA:MSSQLServer-Passw0rd@localhost:11433/master -c "IF NOT EXISTS (SELECT * FROM sys.databases WHERE NAME = 'testdb') CREATE DATABASE testdb;"
	usql ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb -f testdata/mssql.sql
	./testdata/dynamodb.sh > /dev/null 2>&1

test:
	go test ./... -v -coverprofile=coverage.txt -covermode=count
	$(MAKE) testdoc

doc: build
	./tbls doc pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -f sample/postgres
	./tbls doc my://root:mypass@localhost:33306/testdb -c testdata/test_tbls.yml -f sample/mysql
	./tbls doc my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml -f sample/mysql8
	./tbls doc sq://$(PWD)/testdata/testdb.sqlite3 -c testdata/test_tbls.yml -f sample/sqlite
	./tbls doc ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb -c testdata/test_tbls.yml -f sample/mssql
	env AWS_ENDPOINT_URL=http://localhost:18000 ./tbls doc dynamodb://ap-northeast-1 -c testdata/test_tbls_dynamodb.yml -f sample/dynamodb
	./tbls doc pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -j -f sample/adjust
	./tbls doc my://root:mypass@localhost:33306/testdb -c testdata/test_tbls.yml -t svg -f sample/svg
	./tbls doc my://root:mypass@localhost:33306/testdb -c testdata/exclude_test_tbls.yml -f sample/exclude
	./tbls doc my://root:mypass@localhost:33306/testdb -c testdata/dict_test_tbls.yml -f sample/dict

testdoc:
	./tbls diff pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c testdata/test_tbls.yml sample/postgres
	./tbls diff my://root:mypass@localhost:33306/testdb -c testdata/test_tbls.yml sample/mysql
	./tbls diff my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml sample/mysql8
	./tbls diff sq://$(PWD)/testdata/testdb.sqlite3 -c testdata/test_tbls.yml sample/sqlite
	./tbls diff ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb -c testdata/test_tbls.yml sample/mssql
	env AWS_ENDPOINT_URL=http://localhost:18000 ./tbls diff dynamodb://ap-northeast-1 -c testdata/test_tbls_dynamodb.yml sample/dynamodb
	./tbls diff pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c testdata/test_tbls.yml -j sample/adjust
	./tbls diff my://root:mypass@localhost:33306/testdb -c testdata/test_tbls.yml -t svg sample/svg
	./tbls diff my://root:mypass@localhost:33306/testdb -c testdata/exclude_test_tbls.yml sample/exclude
	./tbls diff my://root:mypass@localhost:33306/testdb -c testdata/dict_test_tbls.yml sample/dict

test_too_many_tables:
	usql pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -f testdata/createdb_too_many.sql
	usql pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable -f testdata/createtable_too_many.sql
	ulimit -n 256 && ./tbls doc pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable -f /tmp
	ulimit -n 256 && ./tbls diff pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable /tmp

test_json:
	./tbls out my://root:mypass@localhost:33306/testdb -c testdata/test_tbls.yml -t json > /tmp/tbls.json
	./tbls diff json:///tmp/tbls.json sample/mysql
	./tbls out pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c testdata/test_tbls.yml -t json > /tmp/tbls.json
	./tbls diff json:///tmp/tbls.json sample/postgres
	./tbls out sq://$(PWD)/testdata/testdb.sqlite3 -c testdata/test_tbls.yml -t json > /tmp/tbls.json
	./tbls diff json:///tmp/tbls.json sample/sqlite

test_env:
	env TBLS_DSN=my://root:mypass@localhost:33306/testdb TBLS_DOC_PATH=sample/mysql ./tbls doc -c testdata/test_tbls.yml -f
	env TBLS_DSN=my://root:mypass@localhost:33306/testdb TBLS_DOC_PATH=sample/mysql ./tbls diff -c testdata/test_tbls.yml

test_config:
	./tbls doc -c testdata/mysql_testdb_config.yml -f
	./tbls diff -c testdata/mysql_testdb_config.yml
	cp testdata/mysql_testdb_config.yml .tbls.yml
	./tbls diff
	rm .tbls.yml

doc_bigquery:
	./tbls doc bq://bigquery-public-data/crypto_bitcoin?creds=client_secrets.json -c testdata/crypto_bitcoin_tbls.yml -f sample/bigquery_crypto_bitcoin
	./tbls doc bq://bigquery-public-data/census_bureau_international?creds=client_secrets.json -f sample/bigquery_census_bureau_international

test_bigquery:
	./tbls diff bq://bigquery-public-data/crypto_bitcoin?creds=client_secrets.json -c testdata/crypto_bitcoin_tbls.yml sample/bigquery_crypto_bitcoin
	./tbls diff bq://bigquery-public-data/census_bureau_international?creds=client_secrets.json sample/bigquery_census_bureau_international

doc_spanner:
	./tbls doc spanner://$(GCLOUD_PROJECT)/test-instance/testdb?creds=spanner_client_secrets.json -c testdata/spanner_tbls.yml -f sample/spanner

test_spanner:
	./tbls diff spanner://$(GCLOUD_PROJECT)/test-instance/testdb?creds=spanner_client_secrets.json -c testdata/spanner_tbls.yml sample/spanner

sec:
	gosec ./...

build:
	packr2
	go build -ldflags="$(BUILD_LDFLAGS)"
	packr2 clean

depsdev:
	go get golang.org/x/tools/cmd/cover
	go get golang.org/x/lint/golint
	go get github.com/linyows/git-semv/cmd/git-semv
	go get github.com/Songmu/ghch/cmd/ghch
	go get github.com/xo/usql
	go get github.com/gobuffalo/packr/v2/packr2
	go get github.com/Songmu/gocredits/cmd/gocredits
	go get github.com/securego/gosec/cmd/gosec

prerelease:
	git pull origin --tag
	ghch -w -N ${VER}
	gocredits . > CREDITS
	git add CHANGELOG.md CREDITS
	git commit -m'Bump up version number'
	git tag ${VER}

release:
	goreleaser --rm-dist

.PHONY: default test
