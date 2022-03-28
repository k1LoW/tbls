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

TMPDIR ?= /tmp

default: test

ci: depsdev build db test testdoc testdoc_hide_auto_increment test_too_many_tables test_json test_ext_subcommand sec doc

ci_windows: depsdev build db_sqlite testdoc_sqlite

db: db_sqlite
	usql pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -f testdata/ddl/postgres95.sql
	usql pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -f testdata/ddl/postgres.sql
	usql my://root:mypass@localhost:33306/testdb -f testdata/ddl/mysql56.sql
	usql my://root:mypass@localhost:33308/testdb -c "CREATE DATABASE IF NOT EXISTS relations;"
	usql my://root:mypass@localhost:33308/relations -f testdata/ddl/detect_relations.sql
	usql my://root:mypass@localhost:33308/testdb -c "CREATE DATABASE IF NOT EXISTS relations_singular;"
	usql my://root:mypass@localhost:33308/relations_singular -f testdata/ddl/detect_relations_singular.sql
	usql maria://root:mypass@localhost:33309/testdb -f testdata/ddl/maria.sql
	usql ms://SA:MSSQLServer-Passw0rd@localhost:11433/master -c "IF NOT EXISTS (SELECT * FROM sys.databases WHERE NAME = 'testdb') CREATE DATABASE testdb;"
	usql ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb -f testdata/ddl/mssql.sql || true
	./testdata/ddl/dynamodb.sh > /dev/null 2>&1

db_sqlite:
	sqlite3 $(PWD)/testdata/testdb.sqlite3 < testdata/ddl/sqlite.sql

test:
	go test ./... -v -coverprofile=coverage.out -covermode=count
	$(MAKE) testdoc

doc: build doc_sqlite
	./tbls doc pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -f sample/postgres95
	./tbls doc pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -f sample/postgres
	./tbls doc my://root:mypass@localhost:33306/testdb -c testdata/test_tbls.yml -f sample/mysql56
	./tbls doc my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml -f sample/mysql
	./tbls doc my://root:mypass@localhost:33308/relations -c testdata/test_tbls_detect_relations.yml -f sample/detect_relations
	./tbls doc my://root:mypass@localhost:33308/relations_singular -c testdata/test_tbls_detect_relations_singular.yml -f sample/detect_relations_singular
	./tbls doc maria://root:mypass@localhost:33309/testdb -c testdata/test_tbls.yml -f sample/mariadb
	./tbls doc ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb -c testdata/test_tbls_mssql.yml -f sample/mssql
	env AWS_ENDPOINT_URL=http://localhost:18000 ./tbls doc dynamodb://ap-northeast-1 -c testdata/test_tbls_dynamodb.yml -f sample/dynamodb
	./tbls doc pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -j -f sample/adjust
	./tbls doc my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml -t png -f sample/png
	./tbls doc my://root:mypass@localhost:33308/testdb -c testdata/exclude_test_tbls.yml -f sample/exclude
	./tbls doc my://root:mypass@localhost:33308/testdb -c testdata/dict_test_tbls.yml -f sample/dict
	./tbls doc my://root:mypass@localhost:33308/testdb -c testdata/font_test_tbls.yml -f sample/font
	./tbls doc my://root:mypass@localhost:33308/testdb -c testdata/number_test_tbls.yml -f sample/number

doc_sqlite: build
	./tbls doc sq://$(PWD)/testdata/testdb.sqlite3 -c testdata/test_tbls.yml -f sample/sqlite

testdoc: build testdoc_sqlite
	./tbls diff pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml sample/postgres95
	./tbls diff pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml sample/postgres
	./tbls diff my://root:mypass@localhost:33306/testdb -c testdata/test_tbls.yml sample/mysql56
	./tbls diff my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml sample/mysql
	./tbls diff my://root:mypass@localhost:33308/relations -c testdata/test_tbls_detect_relations.yml sample/detect_relations
	./tbls diff my://root:mypass@localhost:33308/relations_singular -c testdata/test_tbls_detect_relations_singular.yml sample/detect_relations_singular
	./tbls diff maria://root:mypass@localhost:33309/testdb -c testdata/test_tbls.yml sample/mariadb
	./tbls diff ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb -c testdata/test_tbls_mssql.yml sample/mssql
	env AWS_ENDPOINT_URL=http://localhost:18000 ./tbls diff dynamodb://ap-northeast-1 -c testdata/test_tbls_dynamodb.yml sample/dynamodb
	./tbls diff pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -j sample/adjust
	./tbls diff my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml -t png sample/png
	./tbls diff my://root:mypass@localhost:33308/testdb -c testdata/exclude_test_tbls.yml sample/exclude
	./tbls diff my://root:mypass@localhost:33308/testdb -c testdata/dict_test_tbls.yml sample/dict
	./tbls diff my://root:mypass@localhost:33308/testdb -c testdata/font_test_tbls.yml sample/font
	./tbls diff my://root:mypass@localhost:33308/testdb -c testdata/number_test_tbls.yml sample/number

testdoc_sqlite: build
	./tbls diff sq://$(PWD)/testdata/testdb.sqlite3 -c testdata/test_tbls.yml sample/sqlite

testdoc_hide_auto_increment: build
	usql my://root:mypass@localhost:33308/testdb -c "CREATE DATABASE IF NOT EXISTS auto_increment;"
	usql my://root:mypass@localhost:33308/auto_increment -f testdata/ddl/auto_increment.sql
	./tbls doc my://root:mypass@localhost:33308/auto_increment?hide_auto_increment -f $(TMPDIR)/auto_increment
	usql my://root:mypass@localhost:33308/auto_increment -c "INSERT INTO users (username, password, email, created) VALUES ('alice', 'PASS', 'alice@example.com', now());"
	./tbls diff my://root:mypass@localhost:33308/auto_increment?hide_auto_increment $(TMPDIR)/auto_increment

test_too_many_tables: build
	usql pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c "DROP DATABASE IF EXISTS too_many;CREATE DATABASE too_many;"
	usql pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable -f testdata/ddl/createtable_too_many.sql
	ulimit -n 256 && ./tbls doc pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable -f /tmp
	ulimit -n 256 && ./tbls diff pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable /tmp

test_json: build
	./tbls out my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml -t json > /tmp/tbls.json
	./tbls diff json:///tmp/tbls.json sample/mysql
	./tbls out pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -t json > /tmp/tbls.json
	./tbls diff json:///tmp/tbls.json sample/postgres
	./tbls out sq://$(PWD)/testdata/testdb.sqlite3 -c testdata/test_tbls.yml -t json > /tmp/tbls.json
	./tbls diff json:///tmp/tbls.json sample/sqlite

test_env: build
	env TBLS_DSN=my://root:mypass@localhost:33306/testdb TBLS_DOC_PATH=sample/mysql ./tbls doc -c testdata/test_tbls.yml -f
	env TBLS_DSN=my://root:mypass@localhost:33306/testdb TBLS_DOC_PATH=sample/mysql ./tbls diff -c testdata/test_tbls.yml

test_config: build
	./tbls doc -c testdata/mysql_testdb_config.yml -f
	./tbls diff -c testdata/mysql_testdb_config.yml
	cp testdata/mysql_testdb_config.yml .tbls.yml
	./tbls diff
	rm .tbls.yml

doc_bigquery: build
	./tbls doc bq://bigquery-public-data/crypto_bitcoin?creds=client_secrets.json -c testdata/crypto_bitcoin_tbls.yml -f sample/bigquery_crypto_bitcoin
	./tbls doc bq://bigquery-public-data/census_bureau_international?creds=client_secrets.json -f sample/bigquery_census_bureau_international

test_bigquery: build
	./tbls diff bq://bigquery-public-data/crypto_bitcoin?creds=client_secrets.json -c testdata/crypto_bitcoin_tbls.yml sample/bigquery_crypto_bitcoin
	./tbls diff bq://bigquery-public-data/census_bureau_international?creds=client_secrets.json sample/bigquery_census_bureau_international

doc_spanner: build
	./tbls doc spanner://$(GCLOUD_PROJECT)/test-instance/testdb?creds=spanner_client_secrets.json -c testdata/spanner_tbls.yml -f sample/spanner

test_spanner: build
	./tbls diff spanner://$(GCLOUD_PROJECT)/test-instance/testdb?creds=spanner_client_secrets.json -c testdata/spanner_tbls.yml sample/spanner

test_ext_subcommand: build
	echo hello | env PATH="./testdata/bin:${PATH}" ./tbls echo | grep 'STDIN=hello' > /dev/null
	env PATH="./testdata/bin:${PATH}" ./tbls echo -c ./testdata/ext_subcommand_tbls.yml | grep 'TBLS_DSN=pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable' > /dev/null
	env PATH="./testdata/bin:${PATH}" ./tbls echo -c ./testdata/ext_subcommand_tbls.yml | grep 'TBLS_SCHEMA=/' > /dev/null
	env PATH="./testdata/bin:${PATH}" ./tbls echo -c ./testdata/ext_subcommand_tbls.yml | grep 'TBLS_CONFIG_PATH=' | grep 'testdata/ext_subcommand_tbls.yml' > /dev/null
	env PATH="./testdata/bin:${PATH}" TBLS_DSN=pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable ./tbls echo | grep 'TBLS_DSN=pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable' > /dev/null
	echo hello | env PATH="./testdata/bin:${PATH}" ./tbls echo -c ./testdata/ext_subcommand_tbls.yml | grep 'STDIN=hello' > /dev/null

sec:
	gosec ./...

build:
	go build -ldflags="$(BUILD_LDFLAGS)"

depsdev:
	go install github.com/linyows/git-semv/cmd/git-semv@v1.2.0
	go install github.com/Songmu/ghch/cmd/ghch@v0.10.2
	go install github.com/xo/usql@v0.9.5
	go install github.com/Songmu/gocredits/cmd/gocredits@v0.2.0
	go install github.com/securego/gosec/cmd/gosec@master

prerelease:
	git pull origin --tag
	ghch -w -N ${VER}
	gocredits -skip-missing . > CREDITS
	cat _EXTRA_CREDITS >> CREDITS
	git add CHANGELOG.md CREDITS
	git commit -m'Bump up version number'
	git tag ${VER}

release:
	git push origin main --tag
	goreleaser --rm-dist

.PHONY: default test
