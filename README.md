# tbls

`tbls` is a tool for document a database, written in Go.

[Usage](#usage) | [Sample](sample/) | [Integration with CI tools](#integration-with-ci-tools) | [Installation](#installation) | [Database Support](#database-support)

## Usage

### Document a database schema

`tbls doc` analyzes a database and generate document in GitHub Friendly Markdown format.

``` sh
$ tbls doc postgres://user:pass@hostname:5432/dbname ./dbdoc
```

Sample [document](sample/) and [schema](test/pg.sql).

### Diff database schema and document

`tbls diff` shows the difference between database schema and generated document.

``` sh
$ tbls diff postgres://user:pass@hostname:5432/dbname ./dbdoc
```

## Integration with CI tools

1. Commit document using `tbls doc`.
2. Check document updates using `tbls diff`

Set following code to [`your-ci-config.yml`](.travis.yml).

``` sh
DIFF=`./tbls diff postgres://user:pass@localhost:5432/testdb?sslmode=disable ./dbdoc` && if [ ! -z "$DIFF" ]; then echo "document does not updated."; echo $DIFF; exit 1; fi
```

## Installation

``` sh
$ go get github.com/k1LoW/tbls
```

## Database Support

- PostgreSQL
