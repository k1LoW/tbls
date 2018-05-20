# tbls

`tbls` is a tool for analyze database schema. It's one binary, so CI friendly.

## Usage

### Document a database

Use `tbls doc`.

``` sh
$ tbls doc --dsn postgres://user:pass@hostname:3306/dbname --output ./dbdoc
```

Sample [schema](test/pg.sql) and [output](sample/).

## Database driver support

- PostgreSQL
