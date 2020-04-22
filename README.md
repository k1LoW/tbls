<p align="center">
<br>
<img src="https://github.com/k1LoW/tbls/raw/master/img/logo.png" width="200" alt="tbls">
<br><br>
</p>

[![Build Status](https://github.com/k1LoW/tbls/workflows/build/badge.svg)](https://github.com/k1LoW/tbls/actions) [![GitHub release](https://img.shields.io/github/release/k1LoW/tbls.svg)](https://github.com/k1LoW/tbls/releases) [![codecov](https://codecov.io/gh/k1LoW/tbls/branch/master/graph/badge.svg)](https://codecov.io/gh/k1LoW/tbls) [![Go Report Card](https://goreportcard.com/badge/github.com/k1LoW/tbls)](https://goreportcard.com/report/github.com/k1LoW/tbls) [![Docker Build](https://img.shields.io/docker/cloud/build/k1low/tbls)](https://hub.docker.com/r/k1low/tbls)

`tbls` is a CI-Friendly tool for document a database, written in Go.

Key features of `tbls` are:

- **Document a database automatically in [GFM](https://github.github.com/gfm/) format. Output database schema [in many formats](#output-formats).**
- **Single binary = CI-Friendly.**
- **[Support many databases](#support-datasource).**
- **Work as linter for database**

### Table of Contents

  - [Quick Start](#quick-start)
  - [Install](#install)
  - [Getting Started](#getting-started)
    - [Document a database](#document-a-database)
    - [Diff database and document](#diff-database-and-document)
    - [Lint a database](#lint-a-database)
    - [Continuous Integration](#continuous-integration)
  - [Configuration](#configration)
    - [Name](#name)
    - [Description](#description)
    - [Labels](#labels)
    - [DSN](#dsn)
      - [Support Datasource](#support-datasource)
    - [Document path](#document-path)
    - [Table format](#table-format)
    - [ER diagram](#er-diagram)
    - [Filter tables](#filter-tables)
    - [Lint](#lint)
    - [Comments](#comments)
    - [Relations](#relations)
    - [Dictionary](#dictionary)
  - [Output formats](#output-formats)
  - [Command arguments](#command-arguments)
  - [Environment variables](#environment-variables)

<br>

## Quick Start

Document a database with one command.

``` console
$ tbls doc postgres://dbuser:dbpass@hostname:5432/dbname
```

Using docker image.

```console
$ docker run --rm -v $PWD:/work k1low/tbls doc postgres://dbuser:dbpass@hostname:5432/dbname
```

## Install

**homebrew tap:**

```console
$ brew install k1LoW/tap/tbls
```

**manually:**

Download binary from [releases page](https://github.com/k1LoW/tbls/releases)

**go get:**

```console
$ go get github.com/k1LoW/tbls
```

**docker:**

```console
$ docker pull k1low/tbls:latest
```

## Getting Started

### Document a database

Add `.tbls.yml` ( or `tbls.yml` ) file to your repository.

``` yaml
# .tbls.yml

# DSN (Database Source Name) to connect database
dsn: postgres://dbuser:dbpass@localhost:5432/dbname

# Path to generate document
# Default is `dbdoc`
docPath: doc/schema
```

> **Notice:** If you are using a symbol such as `#` `<` in database password, URL-encode the password

Run `tbls doc` to analyzes the database and generate document in GitHub Friendly Markdown format.

``` console
$ tbls doc
```

Commit `.tbls.yml` and the document.

``` console
$ git add .tbls.yml doc/schema
$ git commit -m 'Add database document'
$ git push origin master
```

View the document on GitHub.

[Sample document](sample/postgres/README.md)

![sample](img/doc.png)

### Diff database and document

Update database schema.

``` console
$ psql -U dbuser -d dbname -h hostname -p 5432 -c 'ALTER TABLE users ADD COLUMN phone_number varchar(15);'
Password for user dbuser:
ALTER TABLE
```

`tbls diff` shows the difference between database schema and generated document.

``` diff
$ tbls diff
diff postgres://dbuser:*****@hostname:5432/dbname doc/schema/README.md
--- postgres://dbuser:*****@hostname:5432/dbname
+++ doc/schema/README.md
@@ -4,7 +4,7 @@

 | Name | Columns | Comment | Type |
 | ---- | ------- | ------- | ---- |
-| [users](users.md) | 7 | Users table | BASE TABLE |
+| [users](users.md) | 6 | Users table | BASE TABLE |
 | [user_options](user_options.md) | 4 | User options table | BASE TABLE |
 | [posts](posts.md) | 8 | Posts table | BASE TABLE |
 | [comments](comments.md) | 6 | Comments<br>Multi-line<br>table<br>comment | BASE TABLE |
diff postgres://dbuser:*****@hostname:5432/dbname doc/schema/users.md
--- postgres://dbuser:*****@hostname:5432/dbname
+++ doc/schema/users.md
@@ -14,7 +14,6 @@
 | email | varchar(355) |  | false |  |  | ex. user@example.com |
 | created | timestamp without time zone |  | false |  |  |  |
 | updated | timestamp without time zone |  | true |  |  |  |
-| phone_number | varchar(15) |  | true |  |  |  |

 ## Constraints

```

> **Notice:** `tbls diff` shows the difference Markdown documents only.

### Lint a database

Add linting rule to `.tbls.yml` following

``` yaml
# .tbls.yml
lint:
  requireColumnComment:
    enabled: true
    exclude:
      - id
      - created
      - updated
  columnCount:
    enabled: true
    max: 10
```

Run `tbls lint` to check the database according to `lint:` rules

``` console
$ tbls lint
users.username: column comment required.
users.password: column comment required.
users.phone_number: column comment required.
posts.user_id: column comment required.
posts.title: column comment required.
posts.labels: column comment required.
comments.post_id: column comment required.
comment_stars.user_id: column comment required.
post_comments.comment: column comment required.
posts: too many columns. [12/10]
comments: too many columns. [11/10]

11 detected
```

### Continuous Integration

Continuous integration using tbls.

1. Commit the document using `tbls doc`.
2. Update the database schema in the development cycle.
3. Check for document updates by running `tbls diff` or `tbls lint` in CI.
4. Return to **1**.

**Example: Travis CI**

``` yaml
# .travis.yml
language: go

install:
  - source <(curl -sL https://git.io/use-tbls)
script:
  - tbls diff
  - tbls lint
```

> **Tips:** If your CI based on Debian/Ubuntu (`/bin/sh -> dash`), you can use following install command `curl -sL https://git.io/use-tbls > use-tbls.tmp && . ./use-tbls.tmp && rm ./use-tbls.tmp`

> **Tips:** If the order of the columns does not match, you can use the `--sort` option.

## Configuration

### Name

`name:` is used to specify the database name of the document.

``` yaml
# .tbls.yml
name: mydatabase
```

### Description

`desc:` is used to specify the database description.

``` yaml
# .tbls.yml
desc: This is My Database
```

### Labels

`labels:` is used to label the database or tables.

**label database:**

``` yaml
# .tbls.yml
labels:
  - cmdb
  - analytics
```

**label tables:**

``` yaml
# .tbls.yml
comments:
  -
    table: users
    labels:
      - user
      - privacy data
```

### DSN

`dsn:` (Data Source Name) is used to connect to database.

``` yaml
# .tbls.yml
dsn: my://dbuser:dbpass@hostname:3306/dbname
```

`dsn:` can expand environment variables.

``` yaml
# .tbls.yml
dsn: my://${MYSQL_USER}:${MYSQL_PASSWORD}@hostname:3306/${MYSQL_DATABASE}
```

#### Support Datasource

tbls support following databases/datasources.

**PostgreSQL:**

``` yaml
# .tbls.yml
dsn: postgres://dbuser:dbpass@hostname:5432/dbname
```

``` yaml
# .tbls.yml
dsn: pg://dbuser:dbpass@hostname:5432/dbname
```

**MySQL:**

``` yaml
# .tbls.yml
dsn: mysql://dbuser:dbpass@hostname:3306/dbname
```

``` yaml
# .tbls.yml
dsn: my://dbuser:dbpass@hostname:3306/dbname
```

**SQLite:**

``` yaml
# .tbls.yml
dsn: sqlite:///path/to/dbname.db
```

``` yaml
# .tbls.yml
dsn: sq:///path/to/dbname.db
```

**BigQuery:**

``` yaml
# .tbls.yml
dsn: bigquery://project-id/dataset-id?creds=/path/to/google_application_credentials.json
```

``` yaml
# .tbls.yml
dsn: bq://project-id/dataset-id?creds=/path/to/google_application_credentials.json
```

To set `GOOGLE_APPLICATION_CREDENTIALS` environment variable, you can use

1. `export GOOGLE_APPLICATION_CREDENTIALS` or `env GOOGLE_APPLICATION_CREDENTIALS`
2. Add query to DSN
    - `?google_application_credentials=/path/to/client_secrets.json`
    - `?credentials=/path/to/client_secrets.json`
    - `?creds=/path/to/client_secrets.json`

Required permissions: `bigquery.datasets.get` `bigquery.tables.get` `bigquery.tables.list`

**Cloud Spanner:**

``` yaml
# .tbls.yml
dsn: spanner://project-id/instance-id/dbname?creds=/path/to/google_application_credentials.json
```

To set `GOOGLE_APPLICATION_CREDENTIALS` environment variable, you can use

1. `export GOOGLE_APPLICATION_CREDENTIALS` or `env GOOGLE_APPLICATION_CREDENTIALS`
2. Add query to DSN
    - `?google_application_credentials=/path/to/client_secrets.json`
    - `?credentials=/path/to/client_secrets.json`
    - `?creds=/path/to/client_secrets.json`

**Amazon Redshift:**

``` yaml
# .tbls.yml
dsn: redshift://dbuser:dbpass@hostname:5432/dbname
```

``` yaml
# .tbls.yml
dsn: rs://dbuser:dbpass@hostname:5432/dbname
```

**Microsoft SQL Server:**

``` yaml
# .tbls.yml
dsn: mssql://DbUser:SQLServer-DbPassw0rd@hostname:1433/testdb
```

``` yaml
# .tbls.yml
dsn: sqlserver://DbUser:SQLServer-DbPassw0rd@hostname:1433/testdb
```

``` yaml
# .tbls.yml
dsn: ms://DbUser:SQLServer-DbPassw0rd@localhost:1433/testdb
```

**Amazon DynamoDB:**

``` yaml
# .tbls.yml
dsn: dynamodb://us-west-2
```

``` yaml
# .tbls.yml
dsn: dynamo://ap-northeast-1?aws_access_key_id=XXXXXxxxxxxxXXXXXXX&aws_secret_access_key=XXXXXxxxxxxxXXXXXXX
```

To set AWS credentials, you can use

1. [Use default credential provider chain of AWS SDK for Go](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials)
2. Add query to DSN
    - `?aws_access_key_id=XXXXXxxxxxxxXXXXXXX&aws_secret_access_key=XXXXXxxxxxxxXXXXXXX`

**JSON:**

The JSON file output by the `tbls out -t json` command can be read as a datasource.

``` yaml
---
# .tbls.yml
dsn: json://path/to/testdb.json
```

**HTTP:**

``` yaml
---
# .tbls.yml
dsn: https://hostname/path/to/testdb.json
```

``` yaml
---
# .tbls.yml
dsn:
  url: https://hostname/path/to/testdb.json
  headers:
    Authorization: token GITHUB_OAUTH_TOKEN
```

### Document path

`tbls doc` generates document in the directory specified by `docPath:`.

``` yaml
# .tbls.yml
# Default is `dbdoc`
docPath: doc/schema
```

`docPath:` can expand environment variables.

``` yaml
# .tbls.yml
docPath: ${DOC_PATH}
```

### Table format

`format:` is used to change the document format.

``` yaml
# .tbls.yml
format:
  # Adjust the column width of Markdown format table
  # Default is false
  adjust: true
  # Sort the order of table list and  columns
  # Default is false
  sort: false
```

### ER diagram

`tbls doc` generate ER diagram images at the same time.

``` yaml
# .tbls.yml
er:
  # Skip generation of ER diagram
  # Default is false
  skip: false
  # ER diagram format
  # Default is `png`
  format: svg
  # Add table/column comment to ER diagram
  # Default is false
  comment: true
  # Distance between tables that display relations in the ER
  # Default is 1
  distance: 2
```

### Lint

`tbls lint` work as linter for database.

``` yaml
# .tbls.yml
lint:
  # require table comment
  requireTableComment:
    enabled: true
  # require column comment
  requireColumnComment:
    enabled: true
    # exclude columns from warnings
    exclude:
      - id
      - created_at
      - updated_at
    # exclude tables from warnings
    excludeTables:
      - logs
      - comment_stars
  # find a table that has no relation
  unrelatedTable:
    enabled: true
    # exclude tables from warnings
    exclude:
      - logs
  # check max column count
  columnCount:
    enabled: true
    max: 10
    # exclude tables from warnings
    exclude:
      - user_options
  # require columns
  requireColumns:
    enabled: true
    columns:
      -
        name: created
      -
        name: updated
        exclude:
          - logs
          - CamelizeTable
  # check duplicate relations
  duplicateRelations:
    enabled: true
  # check if the foreign key columns have an index
  requireForeignKeyIndex:
    enabled: true
    exclude:
      - comments.user_id
  # checks if labels are in BigQuery style ( https://cloud.google.com/resource-manager/docs/creating-managing-labels#requirements )
  labelStyleBigQuery:
    enabled: true
    exclude:
      - schema_migrations
```

### Filter tables

![filter tables](img/filter-tables.png)

`include:` and `exclude:` are used to filter target tables from `tbls *`.

``` yaml
# .tbls.yml
include:
  - some_prefix_*
exclude:
  - some_prefix_logs
  - CamelizeTable
```

`lintExclude:` is used to exclude tables from `tbls lint`.

``` yaml
# .tbls.yml
lintExclude:
  - CamelizeTable
```

### Comments

`comments:` is used to add table/column comment to database document without `ALTER TABLE`.

For example, you can add comment about VIEW TABLE or SQLite tables/columns.

``` yaml
# .tbls.yml
comments:
  -
    table: users
    # table comment
    tableComment: Users table
    # column comments
    columnComments:
      email: Email address as login id. ex. user@example.com
  -
    table: post_comments
    tableComment: post and comments View table
    columnComments:
      id: comments.id
      title: posts.title
      post_user: posts.users.username
      comment_user: comments.users.username
      created: comments.created
      updated: comments.updated
```

### Relations

`relations:` is used to add table relation to database document without `FOREIGN KEY`.

You can create ER diagrams with relations without having foreign key constraints.

``` yaml
relations:
  -
    table: logs
    columns:
      - user_id
    parentTable: users
    parentColumns:
      - id
    # Relation definition
    # Default is `Additional Relation`
    def: logs->users
  -
    table: logs
    columns:
      - post_id
    parentTable: posts
    parentColumns:
      - id
  -
    table: logs
    columns:
      - comment_id
    parentTable: comments
    parentColumns:
      - id
  -
    table: logs
    columns:
      - comment_star_id
    parentTable: comment_stars
    parentColumns:
      - id
```

![img](sample/mysql/logs.png)

### Dictionary

`dict:` is used to replace title/table header of database document

``` yaml
# .tbls.yml
---
dict:
  Tables: テーブル一覧
  Description: 概要
  Columns: カラム一覧
  Indexes: INDEX一覧
  Constraints: 制約一覧
  Triggers: トリガー
  Relations: ER図
  Name: 名前
  Comment: コメント
  Type: タイプ
  Default: デフォルト値
  Children: 子テーブル
  Parents: 親テーブル
  Definition: 定義
  Table Definition: テーブル定義
```

## Output formats

`tbls out` output in various formats.

**Markdown:**

``` console
$ tbls out -t md -o schema.md
```

**DOT:**

``` console
$ tbls out -t dot -o schema.dot
```

**PlantUML:**

``` console
$ tbls out -t plantuml -o schema.puml
```

**Image (svg, png, jpg):**

``` console
$ tbls out -t svg --table users --distance 2 -o users.svg
```

**JSON:**

``` console
$ tbls out -t json -o schema.json
```

> **Tips:** `tbls doc` can load `schema.json` as DSN.
>
> ``` console
> $ tbls doc json:///path/to/schema.json
> ```

**YAML:**

``` console
$ tbls out -t yaml -o schema.yaml
```

**Excel:**

``` console
$ tbls out -t xlsx -o schema.xlsx
```

**CSV for [Cacoo's Database Schema Importer](https://support.cacoo.com/hc/en-us/articles/360045672494):**

``` console
$ tbls out -t cacoo -o schema.csv
```

**.tbls.yml:**

``` console
$ tbls out -t config -o .tbls.new.yml
```

## Command arguments

tbls subcommands ( `doc`,`diff`, etc) accepts arguments and options

``` console
$ tbls doc my://root:mypass@localhost:3306/testdb doc/schema
```

You can check available arguments and options using `tbls help [COMMAND]`.

``` console
$ tbls help doc
'tbls doc' analyzes a database and generate document in GitHub Friendly Markdown format.

Usage:
  tbls doc [DSN] [DOC_PATH] [flags]

Flags:
  -a, --add config         additional schema data path (deprecated, use config)
  -j, --adjust-table       adjust column width of table
  -c, --config string      config file path
  -t, --er-format string   ER diagrams output format [png, svg, jpg, ...]. default: png
  -f, --force              force
  -h, --help               help for doc
      --sort               sort
      --without-er         no generate ER diagrams
```

## Environment variables

tbls accepts environment variables `TBLS_DSN` and `TBLS_DOC_PATH`

``` console
$ env TBLS_DSN=my://root:mypass@localhost:3306/testdb TBLS_DOC_PATH=doc/schema tbls doc
```
