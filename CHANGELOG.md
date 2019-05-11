# Changelog

## [v1.12.0](https://github.com/k1LoW/tbls/compare/v1.11.1...v1.12.0) (2019-05-11)

*  Add `exclude` for excluding tables from the document [#96](https://github.com/k1LoW/tbls/pull/96) ([k1LoW](https://github.com/k1LoW))
* Add lint rule `requireColumns` [#95](https://github.com/k1LoW/tbls/pull/95) ([k1LoW](https://github.com/k1LoW))

## [v1.11.1](https://github.com/k1LoW/tbls/compare/v1.11.0...v1.11.1) (2019-04-25)

* Fix loading args when `tbls out` [#91](https://github.com/k1LoW/tbls/pull/91) ([k1LoW](https://github.com/k1LoW))
* Add `--out` option to set output file path. [#90](https://github.com/k1LoW/tbls/pull/90) ([k1LoW](https://github.com/k1LoW))
* Add `xlsx` output format [#89](https://github.com/k1LoW/tbls/pull/89) ([k1LoW](https://github.com/k1LoW))
* Fix .goreleaser.yml for CGO_ENABLED=1 [#88](https://github.com/k1LoW/tbls/pull/88) ([k1LoW](https://github.com/k1LoW))
* Update document [#83](https://github.com/k1LoW/tbls/pull/83) ([k1LoW](https://github.com/k1LoW))
* Update `tbls lint` detect message [#87](https://github.com/k1LoW/tbls/pull/87) ([k1LoW](https://github.com/k1LoW))
* Update `tbls lint` detect message [#87](https://github.com/k1LoW/tbls/pull/87) ([k1LoW](https://github.com/k1LoW))
* Add a temporary installation script for CI [#86](https://github.com/k1LoW/tbls/pull/86) ([k1LoW](https://github.com/k1LoW))
* Add a temporary installation script for CI [#86](https://github.com/k1LoW/tbls/pull/86) ([k1LoW](https://github.com/k1LoW))
* Change `tbls diff` output to unified format [#85](https://github.com/k1LoW/tbls/pull/85) ([k1LoW](https://github.com/k1LoW))
* Change `tbls diff` output to unified format [#85](https://github.com/k1LoW/tbls/pull/85) ([k1LoW](https://github.com/k1LoW))
* Fix config.Config internal [#84](https://github.com/k1LoW/tbls/pull/84) ([k1LoW](https://github.com/k1LoW))
* Fix config.Config internal [#84](https://github.com/k1LoW/tbls/pull/84) ([k1LoW](https://github.com/k1LoW))
* Use goreleaser [#82](https://github.com/k1LoW/tbls/pull/82) ([k1LoW](https://github.com/k1LoW))
* mkdir when document directory does not exists [#81](https://github.com/k1LoW/tbls/pull/81) ([k1LoW](https://github.com/k1LoW))
* Set default doc path `schema` [#80](https://github.com/k1LoW/tbls/pull/80) ([k1LoW](https://github.com/k1LoW))
* Add `md` to `tbls out` format [#79](https://github.com/k1LoW/tbls/pull/79) ([k1LoW](https://github.com/k1LoW))
* Add requireColumnComment.excludedTables [#78](https://github.com/k1LoW/tbls/pull/78) ([k1LoW](https://github.com/k1LoW))
* Fix lint rules [#77](https://github.com/k1LoW/tbls/pull/77) ([k1LoW](https://github.com/k1LoW))
* Fix `dataPath` to `docPath` [#76](https://github.com/k1LoW/tbls/pull/76) ([k1LoW](https://github.com/k1LoW))
* Add `tbls lint` [#75](https://github.com/k1LoW/tbls/pull/75) ([k1LoW](https://github.com/k1LoW))
* Fix the bug that foreign key constraints are not listed in the document and ER diagram in the case of primary key and foreign key [#74](https://github.com/k1LoW/tbls/pull/74) ([k1LoW](https://github.com/k1LoW))
* Support default config file `.tbls.yml` [#73](https://github.com/k1LoW/tbls/pull/73) ([k1LoW](https://github.com/k1LoW))
* `--add` option is deprecated. Use `--config` [#72](https://github.com/k1LoW/tbls/pull/72) ([k1LoW](https://github.com/k1LoW))
* Add `--config` option [#71](https://github.com/k1LoW/tbls/pull/71) ([k1LoW](https://github.com/k1LoW))
* Load Environment Values [#70](https://github.com/k1LoW/tbls/pull/70) ([k1LoW](https://github.com/k1LoW))
* Rename some code [#69](https://github.com/k1LoW/tbls/pull/69) ([k1LoW](https://github.com/k1LoW))
* Add tbls driver information to JSON [#68](https://github.com/k1LoW/tbls/pull/68) ([k1LoW](https://github.com/k1LoW))
* Support `json://`  [#67](https://github.com/k1LoW/tbls/pull/67) ([k1LoW](https://github.com/k1LoW))
* [BREAKING] Change `tbls dot` to `tbls out` / Support JSON format [#66](https://github.com/k1LoW/tbls/pull/66) ([k1LoW](https://github.com/k1LoW))
* Fix error messages [#65](https://github.com/k1LoW/tbls/pull/65) ([k1LoW](https://github.com/k1LoW))
* Fix md output ( fix newline ) [#64](https://github.com/k1LoW/tbls/pull/64) ([k1LoW](https://github.com/k1LoW))
* Add newline at end of file [#63](https://github.com/k1LoW/tbls/pull/63) ([k1LoW](https://github.com/k1LoW))
* Change `tbls diff` exit code [#62](https://github.com/k1LoW/tbls/pull/62) ([k1LoW](https://github.com/k1LoW))
* mv test/ to testdata/ [#61](https://github.com/k1LoW/tbls/pull/61) ([k1LoW](https://github.com/k1LoW))
* Fix `tbls dot` ( use packr ) [#60](https://github.com/k1LoW/tbls/pull/60) ([k1LoW](https://github.com/k1LoW))
* Use Codecov [#59](https://github.com/k1LoW/tbls/pull/59) ([k1LoW](https://github.com/k1LoW))
* Use gobuffalo/packr [#58](https://github.com/k1LoW/tbls/pull/58) ([k1LoW](https://github.com/k1LoW))
* Remove unused modules in go.mod [#57](https://github.com/k1LoW/tbls/pull/57) ([linyows](https://github.com/linyows))
* Support Go 1.11.x [#56](https://github.com/k1LoW/tbls/pull/56) ([k1LoW](https://github.com/k1LoW))
* Fix PostgreSQL constraints sort rule. [#55](https://github.com/k1LoW/tbls/pull/55) ([k1LoW](https://github.com/k1LoW))
* Add `--er-format` option [#54](https://github.com/k1LoW/tbls/pull/54) ([k1LoW](https://github.com/k1LoW))
* Fix relation rendering of multi-columns foreign key [#53](https://github.com/k1LoW/tbls/pull/53) ([k1LoW](https://github.com/k1LoW))
* Support SQLite FTS3/FTS4 Virtual Table [#52](https://github.com/k1LoW/tbls/pull/52) ([k1LoW](https://github.com/k1LoW))
* Fix SQLite relation support [#51](https://github.com/k1LoW/tbls/pull/51) ([k1LoW](https://github.com/k1LoW))
* Support SQLite CHECK constraints [#50](https://github.com/k1LoW/tbls/pull/50) ([k1LoW](https://github.com/k1LoW))
* Support SQLite [#49](https://github.com/k1LoW/tbls/pull/49) ([k1LoW](https://github.com/k1LoW))
* Add Triggers to Schema.Sort() [#47](https://github.com/k1LoW/tbls/pull/47) ([k1LoW](https://github.com/k1LoW))
* Don't show item when length 0 [#46](https://github.com/k1LoW/tbls/pull/46) ([k1LoW](https://github.com/k1LoW))
* Support analyze TRIGGER [#45](https://github.com/k1LoW/tbls/pull/45) ([k1LoW](https://github.com/k1LoW))
* Change option `--no-viz` to `--without-er` [#44](https://github.com/k1LoW/tbls/pull/44) ([k1LoW](https://github.com/k1LoW))
* Fix MySQL constraints / indexes query [#42](https://github.com/k1LoW/tbls/pull/42) ([k1LoW](https://github.com/k1LoW))
* Add exec `dot` STDOUT to error message [#41](https://github.com/k1LoW/tbls/pull/41) ([k1LoW](https://github.com/k1LoW))
*  Add `UNKNOWN CONSTRAINT` ( This is constraint information that "tbls" still can not support ) [#40](https://github.com/k1LoW/tbls/pull/40) ([k1LoW](https://github.com/k1LoW))
* Fix error handling [#39](https://github.com/k1LoW/tbls/pull/39) ([k1LoW](https://github.com/k1LoW))
* Show errors with stack when `DEBUG=1` [#38](https://github.com/k1LoW/tbls/pull/38) ([k1LoW](https://github.com/k1LoW))
* Add `--adjust-table` option for adjust column width of table [#37](https://github.com/k1LoW/tbls/pull/37) ([k1LoW](https://github.com/k1LoW))
* Support CR `\r` and CRLF `\r\n` [#35](https://github.com/k1LoW/tbls/pull/35) ([k1LoW](https://github.com/k1LoW))
* Support multi-line comment [#34](https://github.com/k1LoW/tbls/pull/34) ([k1LoW](https://github.com/k1LoW))
* Fix `too many open files` [#33](https://github.com/k1LoW/tbls/pull/33) ([k1LoW](https://github.com/k1LoW))
* Add test for many tables [#31](https://github.com/k1LoW/tbls/pull/31) ([k1LoW](https://github.com/k1LoW))
* Support PostgreSQL non-default schema [#30](https://github.com/k1LoW/tbls/pull/30) ([k1LoW](https://github.com/k1LoW))
* Work with redshift [#29](https://github.com/k1LoW/tbls/pull/29) ([watarukura](https://github.com/watarukura))
* Fix output dot bug [#25](https://github.com/k1LoW/tbls/pull/25) ([k1LoW](https://github.com/k1LoW))
* Fix table template [#24](https://github.com/k1LoW/tbls/pull/24) ([k1LoW](https://github.com/k1LoW))
* Add schema.Table.Def for show table/view definition [#22](https://github.com/k1LoW/tbls/pull/22) ([k1LoW](https://github.com/k1LoW))
* call dot command with temporary file , graph name in dot file must be quoted [#23](https://github.com/k1LoW/tbls/pull/23) ([kenichiro-kimura](https://github.com/kenichiro-kimura))

## [v1.11.0](https://github.com/k1LoW/tbls/compare/v1.10.1...v1.11.0) (2019-04-25)

* Add `--out` option to set output file path. [#90](https://github.com/k1LoW/tbls/pull/90) ([k1LoW](https://github.com/k1LoW))
* Add `xlsx` output format [#89](https://github.com/k1LoW/tbls/pull/89) ([k1LoW](https://github.com/k1LoW))

## [v1.10.1](https://github.com/k1LoW/tbls/compare/v1.10.0...v1.10.1) (2019-03-16)

* Fix .goreleaser.yml for CGO_ENABLED=1 [#88](https://github.com/k1LoW/tbls/pull/88) ([k1LoW](https://github.com/k1LoW))

## [v1.10.0](https://github.com/k1LoW/tbls/compare/v1.9.0...v1.10.0) (2019-03-13)

* Update document [#83](https://github.com/k1LoW/tbls/pull/83) ([k1LoW](https://github.com/k1LoW))
* Update `tbls lint` detect message [#87](https://github.com/k1LoW/tbls/pull/87) ([k1LoW](https://github.com/k1LoW))
* Update `tbls lint` detect message [#87](https://github.com/k1LoW/tbls/pull/87) ([k1LoW](https://github.com/k1LoW))
* Add a temporary installation script for CI [#86](https://github.com/k1LoW/tbls/pull/86) ([k1LoW](https://github.com/k1LoW))
* Add a temporary installation script for CI [#86](https://github.com/k1LoW/tbls/pull/86) ([k1LoW](https://github.com/k1LoW))
* Change `tbls diff` output to unified format [#85](https://github.com/k1LoW/tbls/pull/85) ([k1LoW](https://github.com/k1LoW))
* Change `tbls diff` output to unified format [#85](https://github.com/k1LoW/tbls/pull/85) ([k1LoW](https://github.com/k1LoW))
* Fix config.Config internal [#84](https://github.com/k1LoW/tbls/pull/84) ([k1LoW](https://github.com/k1LoW))
* Fix config.Config internal [#84](https://github.com/k1LoW/tbls/pull/84) ([k1LoW](https://github.com/k1LoW))

## [v1.9.0](https://github.com/k1LoW/tbls/compare/v1.8.3...v1.9.0) (2019-03-09)

* Use goreleaser [#82](https://github.com/k1LoW/tbls/pull/82) ([k1LoW](https://github.com/k1LoW))
* mkdir when document directory does not exists [#81](https://github.com/k1LoW/tbls/pull/81) ([k1LoW](https://github.com/k1LoW))
* Set default doc path `dbdoc` [#80](https://github.com/k1LoW/tbls/pull/80) [e2ec6ed](https://github.com/k1LoW/tbls/commit/e2ec6ed39016fb80a80f7caeeaefb1821fd100a4) ([k1LoW](https://github.com/k1LoW))
* Add `md` to `tbls out` format [#79](https://github.com/k1LoW/tbls/pull/79) ([k1LoW](https://github.com/k1LoW))

## [v1.8.3](https://github.com/k1LoW/tbls/compare/v1.8.2...v1.8.3) (2019-02-25)

* Add requireColumnComment.excludedTables [#78](https://github.com/k1LoW/tbls/pull/78) ([k1LoW](https://github.com/k1LoW))

## [v1.8.2](https://github.com/k1LoW/tbls/compare/v1.8.1...v1.8.2) (2019-02-25)

* Fix lint rules [#77](https://github.com/k1LoW/tbls/pull/77) ([k1LoW](https://github.com/k1LoW))

## [v1.8.1](https://github.com/k1LoW/tbls/compare/v1.8.0...v1.8.1) (2019-02-25)

* Fix `dataPath` to `docPath` [#76](https://github.com/k1LoW/tbls/pull/76) ([k1LoW](https://github.com/k1LoW))

## [v1.8.0](https://github.com/k1LoW/tbls/compare/v1.7.1...v1.8.0) (2019-02-23)

* Add `tbls lint` [#75](https://github.com/k1LoW/tbls/pull/75) ([k1LoW](https://github.com/k1LoW))

## [v1.7.1](https://github.com/k1LoW/tbls/compare/v1.7.0...v1.7.1) (2018-12-08)

* Fix the bug that foreign key constraints are not listed in the document and ER diagram in the case of primary key and foreign key [#74](https://github.com/k1LoW/tbls/pull/74) ([k1LoW](https://github.com/k1LoW))

## [v1.7.0](https://github.com/k1LoW/tbls/compare/v1.6.0...v1.7.0) (2018-11-29)

* Support default config file `.tbls.yml` [#73](https://github.com/k1LoW/tbls/pull/73) ([k1LoW](https://github.com/k1LoW))

## [v1.6.0](https://github.com/k1LoW/tbls/compare/v1.5.1...v1.6.0) (2018-11-24)

* [DEPRACATED] `--add` option is deprecated. Use `--config` [#72](https://github.com/k1LoW/tbls/pull/72) ([k1LoW](https://github.com/k1LoW))
* Add `--config` option [#71](https://github.com/k1LoW/tbls/pull/71) ([k1LoW](https://github.com/k1LoW))
* Load Environment Values [#70](https://github.com/k1LoW/tbls/pull/70) ([k1LoW](https://github.com/k1LoW))
* Rename some code [#69](https://github.com/k1LoW/tbls/pull/69) ([k1LoW](https://github.com/k1LoW))
* Add tbls driver information to JSON [#68](https://github.com/k1LoW/tbls/pull/68) ([k1LoW](https://github.com/k1LoW))
* Support `json://`  [#67](https://github.com/k1LoW/tbls/pull/67) ([k1LoW](https://github.com/k1LoW))
* [BREAKING] Change `tbls dot` to `tbls out` / Support JSON format [#66](https://github.com/k1LoW/tbls/pull/66) ([k1LoW](https://github.com/k1LoW))

## [v1.5.1](https://github.com/k1LoW/tbls/compare/v1.5.0...v1.5.1) (2018-11-20)

* Fix error messages [#65](https://github.com/k1LoW/tbls/pull/65) ([k1LoW](https://github.com/k1LoW))

## [v1.5.0](https://github.com/k1LoW/tbls/compare/v1.4.1...v1.5.0) (2018-11-18)

* [BREAKING] Fix md output ( fix newline ) [#64](https://github.com/k1LoW/tbls/pull/64) ([k1LoW](https://github.com/k1LoW))
* [BREAKING] Add newline at end of file [#63](https://github.com/k1LoW/tbls/pull/63) ([k1LoW](https://github.com/k1LoW))
* [BREAKING] Change `tbls diff` exit code [#62](https://github.com/k1LoW/tbls/pull/62) ([k1LoW](https://github.com/k1LoW))
* mv test/ to testdata/ [#61](https://github.com/k1LoW/tbls/pull/61) ([k1LoW](https://github.com/k1LoW))
* Fix `tbls dot` ( use packr ) [#60](https://github.com/k1LoW/tbls/pull/60) ([k1LoW](https://github.com/k1LoW))
* Use Codecov [#59](https://github.com/k1LoW/tbls/pull/59) ([k1LoW](https://github.com/k1LoW))

## [v1.4.1](https://github.com/k1LoW/tbls/compare/v1.4.0...v1.4.1) (2018-11-14)

* Use gobuffalo/packr [#58](https://github.com/k1LoW/tbls/pull/58) ([k1LoW](https://github.com/k1LoW))
* Remove unused modules in go.mod [#57](https://github.com/k1LoW/tbls/pull/57) ([linyows](https://github.com/linyows))

## [v1.4.0](https://github.com/k1LoW/tbls/compare/v1.3.0...v1.4.0) (2018-11-13)

* Support Go 1.11.x [#56](https://github.com/k1LoW/tbls/pull/56) ([k1LoW](https://github.com/k1LoW))
* Fix PostgreSQL constraints sort rule. [#55](https://github.com/k1LoW/tbls/pull/55) ([k1LoW](https://github.com/k1LoW))

## [v1.3.0](https://github.com/k1LoW/tbls/compare/v1.2.1...v1.3.0) (2018-09-06)

* Add `--er-format` option [#54](https://github.com/k1LoW/tbls/pull/54) ([k1LoW](https://github.com/k1LoW))

## [v1.2.1](https://github.com/k1LoW/tbls/compare/v1.2.0...v1.2.1) (2018-08-09)

* Fix relation rendering of multi-columns foreign key [#53](https://github.com/k1LoW/tbls/pull/53) ([k1LoW](https://github.com/k1LoW))

## [v1.2.0](https://github.com/k1LoW/tbls/compare/v1.1.1...v1.2.0) (2018-08-08)

* Support SQLite FTS3/FTS4 Virtual Table [#52](https://github.com/k1LoW/tbls/pull/52) ([k1LoW](https://github.com/k1LoW))
* Fix SQLite relation support [#51](https://github.com/k1LoW/tbls/pull/51) ([k1LoW](https://github.com/k1LoW))

## [v1.1.1](https://github.com/k1LoW/tbls/compare/v1.1.0...v1.1.1) (2018-08-06)

* Support SQLite CHECK constraints [#50](https://github.com/k1LoW/tbls/pull/50) ([k1LoW](https://github.com/k1LoW))

## [v1.1.0](https://github.com/k1LoW/tbls/compare/v1.0.1...v1.1.0) (2018-08-05)

* Support SQLite [#49](https://github.com/k1LoW/tbls/pull/49) ([k1LoW](https://github.com/k1LoW))

## [v1.0.1](https://github.com/k1LoW/tbls/compare/v1.0.0...v1.0.1) (2018-07-28)

* Add Triggers to Schema.Sort() [#47](https://github.com/k1LoW/tbls/pull/47) ([k1LoW](https://github.com/k1LoW))

## [v1.0.0](https://github.com/k1LoW/tbls/compare/v0.10.2...v1.0.0) (2018-07-28)

* Don't show item when length 0 [#46](https://github.com/k1LoW/tbls/pull/46) ([k1LoW](https://github.com/k1LoW))
* Support analyze TRIGGER [#45](https://github.com/k1LoW/tbls/pull/45) ([k1LoW](https://github.com/k1LoW))
* Change option `--no-viz` to `--without-er` [#44](https://github.com/k1LoW/tbls/pull/44) ([k1LoW](https://github.com/k1LoW))

## [v0.10.2](https://github.com/k1LoW/tbls/compare/v0.10.1...v0.10.2) (2018-07-26)

* Fix MySQL constraints / indexes query [#42](https://github.com/k1LoW/tbls/pull/42) ([k1LoW](https://github.com/k1LoW))
* Add exec `dot` STDOUT to error message [#41](https://github.com/k1LoW/tbls/pull/41) ([k1LoW](https://github.com/k1LoW))
*  Add `UNKNOWN CONSTRAINT` ( This is constraint information that "tbls" still can not support ) [#40](https://github.com/k1LoW/tbls/pull/40) ([k1LoW](https://github.com/k1LoW))

## [v0.10.1](https://github.com/k1LoW/tbls/compare/v0.10.0...v0.10.1) (2018-07-22)

* Fix error handling [#39](https://github.com/k1LoW/tbls/pull/39) ([k1LoW](https://github.com/k1LoW))

## [v0.10.0](https://github.com/k1LoW/tbls/compare/v0.9.3...v0.10.0) (2018-07-22)

* Show errors with stack when `DEBUG=1` [#38](https://github.com/k1LoW/tbls/pull/38) ([k1LoW](https://github.com/k1LoW))
* Add `--adjust-table` option for adjust column width of table [#37](https://github.com/k1LoW/tbls/pull/37) ([k1LoW](https://github.com/k1LoW))

## [v0.9.3](https://github.com/k1LoW/tbls/compare/v0.9.2...v0.9.3) (2018-07-13)

* Support CR `\r` and CRLF `\r\n` [#35](https://github.com/k1LoW/tbls/pull/35) ([k1LoW](https://github.com/k1LoW))

## [v0.9.2](https://github.com/k1LoW/tbls/compare/v0.9.1...v0.9.2) (2018-07-09)

* Support multi-line comment [#34](https://github.com/k1LoW/tbls/pull/34) ([k1LoW](https://github.com/k1LoW))
* Fix `too many open files` [#33](https://github.com/k1LoW/tbls/pull/33) ([k1LoW](https://github.com/k1LoW))
* Add test for many tables [#31](https://github.com/k1LoW/tbls/pull/31) ([k1LoW](https://github.com/k1LoW))

## [v0.9.1](https://github.com/k1LoW/tbls/compare/v0.9.0...v0.9.1) (2018-06-30)

* Support PostgreSQL non-default schema [#30](https://github.com/k1LoW/tbls/pull/30) ([k1LoW](https://github.com/k1LoW))

## [v0.9.0](https://github.com/k1LoW/tbls/compare/v0.8.2...v0.9.0) (2018-06-29)

* Work with Amazon Redshift [#29](https://github.com/k1LoW/tbls/pull/29) ([watarukura](https://github.com/watarukura))

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
