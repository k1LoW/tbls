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

TBLS ?= ./tbls

default: test

ci: depsdev build db test testdoc testdoc_hide_auto_increment test_too_many_tables test_json test_ext_subcommand doc

ci_windows: depsdev build db_sqlite testdoc_sqlite

db: db_sqlite # MySQL8 use ./testdata/ddl/mysql:/docker-entrypoint-initdb.d
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
	go test ./... -tags 'bq clickhouse dynamo mariadb mongodb mssql mysql postgres redshift snowflake spanner sqlite' -coverprofile=coverage.out -covermode=count

test-no-db:
	go test ./... -coverprofile=coverage.out -covermode=count

doc: build doc_sqlite
	$(TBLS) doc pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -f sample/postgres95
	$(TBLS) doc pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -f sample/postgres
	$(TBLS) doc my://root:mypass@localhost:33306/testdb -c testdata/test_tbls.yml -f sample/mysql56
	$(TBLS) doc my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml -f sample/mysql
	$(TBLS) doc my://root:mypass@localhost:33308/relations -c testdata/test_tbls_detect_relations.yml -f sample/detect_relations
	$(TBLS) doc my://root:mypass@localhost:33308/relations_singular -c testdata/test_tbls_detect_relations_singular.yml -f sample/detect_relations_singular
	$(TBLS) doc maria://root:mypass@localhost:33309/testdb -c testdata/test_tbls.yml -f sample/mariadb
	$(TBLS) doc ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb -c testdata/test_tbls_mssql.yml -f sample/mssql
	$(TBLS) doc mongodb://mongoadmin:secret@localhost:27017/test?authSource=admin -f sample/mongo
	env AWS_ENDPOINT_URL=http://localhost:18000 $(TBLS) doc dynamodb://ap-northeast-1 -c testdata/test_tbls_dynamodb.yml -f sample/dynamodb
	$(TBLS) doc clickhouse://default@localhost:9000/testdb -f sample/clickhouse
	$(TBLS) doc pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -j -f sample/adjust
	$(TBLS) doc my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml -t png -f sample/png
	$(TBLS) doc my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml -t mermaid -f sample/mermaid
	$(TBLS) doc my://root:mypass@localhost:33308/testdb -c testdata/exclude_test_tbls.yml -f sample/exclude
	$(TBLS) doc my://root:mypass@localhost:33308/testdb -c testdata/dict_test_tbls.yml -f sample/dict
	$(TBLS) doc my://root:mypass@localhost:33308/testdb -c testdata/font_test_tbls.yml -f sample/font
	$(TBLS) doc my://root:mypass@localhost:33308/testdb -c testdata/number_test_tbls.yml -f sample/number
	$(TBLS) doc my://root:mypass@localhost:33308/testdb -c testdata/hide_test_tbls.yml -f sample/hide
	$(TBLS) doc my://root:mypass@localhost:33308/testdb -c testdata/test_tbls_hide_not_related_column.yml -f sample/hide_not_related_column
	$(TBLS) doc my://root:mypass@localhost:33308/testdb -c testdata/test_tbls_viewpoints.yml -f sample/viewpoints

doc_sqlite: build
	$(TBLS) doc sq://$(PWD)/testdata/testdb.sqlite3 -c testdata/test_tbls.yml -f sample/sqlite

testdoc: build testdoc_sqlite
	$(TBLS) diff pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml sample/postgres95
	$(TBLS) diff pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml sample/postgres
	$(TBLS) diff my://root:mypass@localhost:33306/testdb -c testdata/test_tbls.yml sample/mysql56
	$(TBLS) diff my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml sample/mysql
	$(TBLS) diff my://root:mypass@localhost:33308/relations -c testdata/test_tbls_detect_relations.yml sample/detect_relations
	$(TBLS) diff my://root:mypass@localhost:33308/relations_singular -c testdata/test_tbls_detect_relations_singular.yml sample/detect_relations_singular
	$(TBLS) diff maria://root:mypass@localhost:33309/testdb -c testdata/test_tbls.yml sample/mariadb
	$(TBLS) diff ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb -c testdata/test_tbls_mssql.yml sample/mssql
	$(TBLS) diff mongodb://mongoadmin:secret@localhost:27017/test?authSource=admin sample/mongo
	env AWS_ENDPOINT_URL=http://localhost:18000 $(TBLS) diff dynamodb://ap-northeast-1 -c testdata/test_tbls_dynamodb.yml sample/dynamodb
	$(TBLS) diff clickhouse://default@localhost:9000/testdb sample/clickhouse
	$(TBLS) diff pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -j sample/adjust
	$(TBLS) diff my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml -t png sample/png
	$(TBLS) diff my://root:mypass@localhost:33308/testdb -c testdata/exclude_test_tbls.yml sample/exclude
	$(TBLS) diff my://root:mypass@localhost:33308/testdb -c testdata/dict_test_tbls.yml sample/dict
	$(TBLS) diff my://root:mypass@localhost:33308/testdb -c testdata/font_test_tbls.yml sample/font
	$(TBLS) diff my://root:mypass@localhost:33308/testdb -c testdata/number_test_tbls.yml sample/number
	$(TBLS) diff my://root:mypass@localhost:33308/testdb -c testdata/hide_test_tbls.yml sample/hide
	$(TBLS) diff my://root:mypass@localhost:33308/testdb -c testdata/test_tbls_hide_not_related_column.yml sample/hide_not_related_column
	$(TBLS) diff my://root:mypass@localhost:33308/testdb -c testdata/test_tbls_viewpoints.yml sample/viewpoints

testdoc_sqlite: build
	$(TBLS) diff sq://$(PWD)/testdata/testdb.sqlite3 -c testdata/test_tbls.yml sample/sqlite

testdoc_hide_auto_increment: build
	usql my://root:mypass@localhost:33308/testdb -c "CREATE DATABASE IF NOT EXISTS auto_increment;"
	usql my://root:mypass@localhost:33308/auto_increment -f testdata/ddl/auto_increment.sql
	$(TBLS) doc my://root:mypass@localhost:33308/auto_increment?hide_auto_increment -f $(TMPDIR)/auto_increment
	usql my://root:mypass@localhost:33308/auto_increment -c "INSERT INTO users (username, password, email, created) VALUES ('alice', 'PASS', 'alice@example.com', now());"
	$(TBLS) diff my://root:mypass@localhost:33308/auto_increment?hide_auto_increment $(TMPDIR)/auto_increment

test_too_many_tables: build
	usql pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable -c "DROP DATABASE IF EXISTS too_many;CREATE DATABASE too_many;"
	usql pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable -f testdata/ddl/createtable_too_many.sql
	ulimit -n 256 && $(TBLS) doc pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable -f /tmp
	ulimit -n 256 && $(TBLS) diff pg://postgres:pgpass@localhost:55432/too_many?sslmode=disable /tmp

test_json: build
	$(TBLS) out my://root:mypass@localhost:33308/testdb -c testdata/test_tbls.yml -t json > /tmp/tbls.json
	$(TBLS) diff json:///tmp/tbls.json sample/mysql
	$(TBLS) out pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable -c testdata/test_tbls_postgres.yml -t json > /tmp/tbls.json
	$(TBLS) diff json:///tmp/tbls.json sample/postgres
	$(TBLS) out sq://$(PWD)/testdata/testdb.sqlite3 -c testdata/test_tbls.yml -t json > /tmp/tbls.json
	$(TBLS) diff json:///tmp/tbls.json sample/sqlite

test_env: build
	env TBLS_DSN=my://root:mypass@localhost:33306/testdb TBLS_DOC_PATH=sample/mysql $(TBLS) doc -c testdata/test_tbls.yml -f
	env TBLS_DSN=my://root:mypass@localhost:33306/testdb TBLS_DOC_PATH=sample/mysql $(TBLS) diff -c testdata/test_tbls.yml

test_config: build
	$(TBLS) doc -c testdata/mysql_testdb_config.yml -f
	$(TBLS) diff -c testdata/mysql_testdb_config.yml
	cp testdata/mysql_testdb_config.yml .tbls.yml
	$(TBLS) diff
	rm .tbls.yml

doc_bigquery: build
	$(TBLS) doc bq://bigquery-public-data/crypto_bitcoin?creds=client_secrets.json -c testdata/crypto_bitcoin_tbls.yml -f sample/bigquery_crypto_bitcoin
	$(TBLS) doc bq://bigquery-public-data/census_bureau_international?creds=client_secrets.json -f sample/bigquery_census_bureau_international

test_bigquery: build
	$(TBLS) diff bq://bigquery-public-data/crypto_bitcoin?creds=client_secrets.json -c testdata/crypto_bitcoin_tbls.yml sample/bigquery_crypto_bitcoin
	$(TBLS) diff bq://bigquery-public-data/census_bureau_international?creds=client_secrets.json sample/bigquery_census_bureau_international

doc_spanner: build
	$(TBLS) doc spanner://$(GCLOUD_PROJECT)/test-instance/testdb?creds=spanner_client_secrets.json -c testdata/spanner_tbls.yml -f sample/spanner

test_spanner: build
	$(TBLS) diff spanner://$(GCLOUD_PROJECT)/test-instance/testdb?creds=spanner_client_secrets.json -c testdata/spanner_tbls.yml sample/spanner

test_ext_subcommand: build
	echo hello | env PATH="${PWD}/testdata/bin:${PATH}" $(TBLS) echo | grep 'STDIN=hello' > /dev/null
	env PATH="${PWD}/testdata/bin:${PATH}" $(TBLS) echo -c ./testdata/ext_subcommand_tbls.yml | grep 'TBLS_DSN=pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable' > /dev/null
	env PATH="${PWD}/testdata/bin:${PATH}" $(TBLS) echo -c ./testdata/ext_subcommand_tbls.yml | grep 'TBLS_SCHEMA=/' > /dev/null
	env PATH="${PWD}/testdata/bin:${PATH}" $(TBLS) echo -c ./testdata/ext_subcommand_tbls.yml | grep 'TBLS_CONFIG_PATH=' | grep 'testdata/ext_subcommand_tbls.yml' > /dev/null
	env PATH="${PWD}/testdata/bin:${PATH}" TBLS_DSN=pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable $(TBLS) echo | grep 'TBLS_DSN=pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable' > /dev/null
	echo hello | env PATH="${PWD}/testdata/bin:${PATH}" $(TBLS) echo -c ./testdata/ext_subcommand_tbls.yml | grep 'STDIN=hello' > /dev/null

generate_test_json: build
	sqlite3 $(PWD)/filter_tables.sqlite3 < testdata/ddl/filter_tables.sql
	$(TBLS) out sq://$(PWD)/filter_tables.sqlite3 -t json > testdata/filter_tables.json

lint:
	golangci-lint run ./...

build:
	go build -tags timetzdata -ldflags="$(BUILD_LDFLAGS)"

depsdev:
	go install github.com/linyows/git-semv/cmd/git-semv@v1.2.0
	go install github.com/Songmu/ghch/cmd/ghch@v0.10.2
	go install github.com/xo/usql@v0.9.5
	go install github.com/Songmu/gocredits/cmd/gocredits@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

prerelease:
	git pull origin --tag
	ghch -w -N ${VER}
	gocredits -w .
	git add CHANGELOG.md CREDITS
	git commit -m'Bump up version number'
	git tag ${VER}

prerelease_for_tagpr: depsdev
	gocredits -w .
	git add CHANGELOG.md CREDITS go.mod go.sum

.PHONY: default test
