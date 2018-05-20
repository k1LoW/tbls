# tbls

`tbls` is a tool for document a database, written in Go.

[Usage](#usage) | [Sample](sample/) | [Install](#install) | [Database Support](#database-support)

## Usage

### Document a database schema

`tbls doc` analyzes a database and generate document in GitHub Friendly Markdown format.

``` sh
$ tbls doc postgres://user:pass@hostname:3306/dbname ./dbdoc
```

Sample [document](sample/) and [schema](test/pg.sql).

### Diff database schema and document

`tbls diff` shows the difference between database schema and generated document.

``` sh
$ tbls diff postgres://user:pass@hostname:3306/dbname ./dbdoc
```

## Install

``` sh
$ go get github.com/k1LoW/tbls
```

## Database Support

- PostgreSQL
