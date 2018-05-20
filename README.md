# tbls

`tbls` is a tool for analyze database schema, written in Go.

## Usage

### Document a database schema

Use `tbls doc`.

``` sh
$ tbls doc postgres://user:pass@hostname:3306/dbname ./dbdoc
```

Sample [schema](test/pg.sql) and [documented markdown files](sample/).

### Diff database schema and document

Use `tbls diff`.

``` sh
$ tbls diff postgres://user:pass@hostname:3306/dbname ./dbdoc
```

## Database support

- PostgreSQL
