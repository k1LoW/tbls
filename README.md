# tbls

`tbls` is a tool for analyze database schema, written in Go.

## Usage

### Document a database schema

`tbls doc` analyzes a database and generate document in GitHub Friendly Markdown format.

``` sh
$ tbls doc postgres://user:pass@hostname:3306/dbname ./dbdoc
```

Sample [schema](test/pg.sql) and [documented markdown files](sample/).

### Diff database schema and document

`tbls diff` shows the difference between database schema and generated document.

``` sh
$ tbls diff postgres://user:pass@hostname:3306/dbname ./dbdoc
```

## Database support

- PostgreSQL
