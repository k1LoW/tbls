# Changelog

## [v0.8.2](https://github.com/k1LoW/tbls/compare/v0.8.1...v0.8.2) (2018-06-06)

* Fix output dot bug [#25](https://github.com/k1LoW/tbls/pull/25) ([k1LoW](https://github.com/k1LoW))

## [v0.8.1](https://github.com/k1LoW/tbls/compare/v0.8.0...v0.8.1) (2018-06-06)

* Fix table template [#24](https://github.com/k1LoW/tbls/pull/24) ([k1LoW](https://github.com/k1LoW))

## [v0.8.0](https://github.com/k1LoW/tbls/compare/v0.7.0...v0.8.0) (2018-06-05)

* Add schema.Table.Def for show table/view definition [#22](https://github.com/k1LoW/tbls/pull/22) ([k1LoW](https://github.com/k1LoW))
* call dot command with temporary file , graph name in dot file must be quoted [#23](https://github.com/k1LoW/tbls/pull/23) ([kenichiro-kimura](https://github.com/kenichiro-kimura))

## [v0.7.0](https://github.com/k1LoW/tbls/compare/v0.6.2...v0.7.0) (2018-06-02)

* `--add` option support additional comments [#21](https://github.com/k1LoW/tbls/pull/21) ([k1LoW](https://github.com/k1LoW))

## [v0.6.2](https://github.com/k1LoW/tbls/compare/v0.6.1...v0.6.2) (2018-05-31)

* Add `ORDER BY` to sort columns, constraints [#20](https://github.com/k1LoW/tbls/pull/20) ([k1LoW](https://github.com/k1LoW))

## [v0.6.1](https://github.com/k1LoW/tbls/compare/v0.6.0...v0.6.1) (2018-05-30)

* Escape tmpl value because dot file use <TABLE> [#19](https://github.com/k1LoW/tbls/pull/19) ([k1LoW](https://github.com/k1LoW))
* Change style of additional relation edges / can set relation def [#18](https://github.com/k1LoW/tbls/pull/18) ([k1LoW](https://github.com/k1LoW))

## [v0.6.0](https://github.com/k1LoW/tbls/compare/v0.5.1...v0.6.0) (2018-05-30)

* Use Graphviz `dot` to generate ER diagram .png [#17](https://github.com/k1LoW/tbls/pull/17) ([k1LoW](https://github.com/k1LoW))
* `tbls dot` generate dot to STDOUT [#15](https://github.com/k1LoW/tbls/pull/15) ([k1LoW](https://github.com/k1LoW))

## [v0.5.1](https://github.com/k1LoW/tbls/compare/v0.5.0...v0.5.1) (2018-05-29)

* Support Camelize table name like "CamelizeTable" [#16](https://github.com/k1LoW/tbls/pull/16) ([k1LoW](https://github.com/k1LoW))

## [v0.5.0](https://github.com/k1LoW/tbls/compare/v0.4.0...v0.5.0) (2018-05-28)

* Support CHECK constraints [#14](https://github.com/k1LoW/tbls/pull/14) ([k1LoW](https://github.com/k1LoW))
* Support view table [#13](https://github.com/k1LoW/tbls/pull/13) ([k1LoW](https://github.com/k1LoW))
* Add `tbls dot` command [#12](https://github.com/k1LoW/tbls/pull/12) ([k1LoW](https://github.com/k1LoW))

## [v0.4.0](https://github.com/k1LoW/tbls/compare/v0.3.0...v0.4.0) (2018-05-26)

* Add `--add` option for add extra data (relations) to schema [#11](https://github.com/k1LoW/tbls/pull/11) ([k1LoW](https://github.com/k1LoW))
* Support MySQL 8 [#10](https://github.com/k1LoW/tbls/pull/10) ([k1LoW](https://github.com/k1LoW))
* Add db.Ping() for test connection [#9](https://github.com/k1LoW/tbls/pull/9) ([k1LoW](https://github.com/k1LoW))

## [v0.3.0](https://github.com/k1LoW/tbls/compare/v0.2.2...v0.3.0) (2018-05-24)

* MySQL driver support [#8](https://github.com/k1LoW/tbls/pull/8) ([k1LoW](https://github.com/k1LoW))

## [v0.2.2](https://github.com/k1LoW/tbls/compare/v0.2.1...v0.2.2) (2018-05-24)

* Fix typo ;; [#7](https://github.com/k1LoW/tbls/pull/7) ([k1LoW](https://github.com/k1LoW))

## [v0.2.1](https://github.com/k1LoW/tbls/compare/v0.2.0...v0.2.1) (2018-05-22)

* Fix query for tables [#6](https://github.com/k1LoW/tbls/pull/6) ([k1LoW](https://github.com/k1LoW))

## [v0.2.0](https://github.com/k1LoW/tbls/compare/v0.1.2...v0.2.0) (2018-05-21)

* Add `--sort` option for CI easily [#5](https://github.com/k1LoW/tbls/pull/5) ([k1LoW](https://github.com/k1LoW))
* Add go-assets [#4](https://github.com/k1LoW/tbls/pull/4) ([k1LoW](https://github.com/k1LoW))

## [v0.1.2](https://github.com/k1LoW/tbls/compare/v0.1.1...v0.1.2) (2018-05-21)

* Fix defer *.Close() [#3](https://github.com/k1LoW/tbls/pull/3) ([k1LoW](https://github.com/k1LoW))
* Use path/filepath [#2](https://github.com/k1LoW/tbls/pull/2) ([k1LoW](https://github.com/k1LoW))

## [v0.1.1](https://github.com/k1LoW/tbls/compare/131f7e2eb87f...v0.1.1) (2018-05-21)

* Add driver interface [#1](https://github.com/k1LoW/tbls/pull/1) ([k1LoW](https://github.com/k1LoW))

## [v0.1.0](https://github.com/k1LoW/tbls/compare/131f7e2eb87f...v0.1.0) (2018-05-20)
