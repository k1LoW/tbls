# Changelog

## [v1.78.1](https://github.com/k1LoW/tbls/compare/v1.78.0...v1.78.1) - 2024-10-18
### Other Changes
- feat(postgres): Ensure ordering of index columns is always stable by @95ulisse in https://github.com/k1LoW/tbls/pull/623

## [v1.78.0](https://github.com/k1LoW/tbls/compare/v1.77.0...v1.78.0) - 2024-10-08
### Breaking Changes 🛠
- Use github.com/k1LoW/errors by @k1LoW in https://github.com/k1LoW/tbls/pull/613
- Add support to output enum definitions in the db doc README by @insano10 in https://github.com/k1LoW/tbls/pull/601
### Fix bug 🐛
- Fix CI by @k1LoW in https://github.com/k1LoW/tbls/pull/621
### Other Changes
- Fix using docker by @k1LoW in https://github.com/k1LoW/tbls/pull/614
- chore(deps): bump the dependencies group across 1 directory with 8 updates by @dependabot in https://github.com/k1LoW/tbls/pull/616
- chore(deps): bump the dependencies group with 6 updates by @dependabot in https://github.com/k1LoW/tbls/pull/618
- chore(deps): bump the dependencies group with 8 updates by @dependabot in https://github.com/k1LoW/tbls/pull/620
- Add existing virtual relation strategies to the README by @suzuki in https://github.com/k1LoW/tbls/pull/622

## [v1.77.0](https://github.com/k1LoW/tbls/compare/v1.76.1...v1.77.0) - 2024-07-24
### Breaking Changes 🛠
- Refactor when expression handling by @mjpieters in https://github.com/k1LoW/tbls/pull/600
- Hide inner table (ClickHouse) by @k1LoW in https://github.com/k1LoW/tbls/pull/610
### New Features 🎉
- Add support for ClickHouse by @joschi in https://github.com/k1LoW/tbls/pull/605
### Other Changes
- Update .goreleaser.yml by @k1LoW in https://github.com/k1LoW/tbls/pull/597
- chore(deps): bump golang.org/x/image from 0.16.0 to 0.18.0 by @dependabot in https://github.com/k1LoW/tbls/pull/602
- chore(deps): bump docker/build-push-action from 5 to 6 in the dependencies group by @dependabot in https://github.com/k1LoW/tbls/pull/603
- chore(deps): bump google.golang.org/grpc from 1.64.0 to 1.64.1 by @dependabot in https://github.com/k1LoW/tbls/pull/606
- Add ClickHouse support to README by @joschi in https://github.com/k1LoW/tbls/pull/608
- Add ClickHouse sample by @k1LoW in https://github.com/k1LoW/tbls/pull/609
- chore(deps): bump the dependencies group across 1 directory with 11 updates by @dependabot in https://github.com/k1LoW/tbls/pull/611

## [v1.76.1](https://github.com/k1LoW/tbls/compare/v1.76.0...v1.76.1) - 2024-06-19
### New Features 🎉
- Postgres: Support "$user" search path by @mjpieters in https://github.com/k1LoW/tbls/pull/594
### Other Changes
- Address lint errors by @mjpieters in https://github.com/k1LoW/tbls/pull/595

## [v1.76.0](https://github.com/k1LoW/tbls/compare/v1.75.0...v1.76.0) - 2024-06-04
### Breaking Changes 🛠
- Show table comments for related tables of View (#587) by @majimaccho in https://github.com/k1LoW/tbls/pull/590
### Other Changes
- chore(deps): bump the dependencies group with 10 updates by @dependabot in https://github.com/k1LoW/tbls/pull/588

## [v1.75.0](https://github.com/k1LoW/tbls/compare/v1.74.4...v1.75.0) - 2024-05-16
### Breaking Changes 🛠
- Embed tzdata by @k1LoW in https://github.com/k1LoW/tbls/pull/585

## [v1.74.4](https://github.com/k1LoW/tbls/compare/v1.74.3...v1.74.4) - 2024-05-16
### Other Changes
- Fix some minor typos by @fkmy in https://github.com/k1LoW/tbls/pull/583

## [v1.74.3](https://github.com/k1LoW/tbls/compare/v1.74.2...v1.74.3) - 2024-05-06
### Other Changes
- Use `ghfs` for AnalyzeGithubContent by @kromiii in https://github.com/k1LoW/tbls/pull/581

## [v1.74.2](https://github.com/k1LoW/tbls/compare/v1.74.1...v1.74.2) - 2024-05-02
### Fix bug 🐛
- Support the case where name in index_info is NULL. by @k1LoW in https://github.com/k1LoW/tbls/pull/579

## [v1.74.1](https://github.com/k1LoW/tbls/compare/v1.74.0...v1.74.1) - 2024-05-01
### Other Changes
- chore(deps): bump golang.org/x/net from 0.22.0 to 0.23.0 by @dependabot in https://github.com/k1LoW/tbls/pull/575
- chore(deps): bump the dependencies group with 7 updates by @dependabot in https://github.com/k1LoW/tbls/pull/577

## [v1.74.0](https://github.com/k1LoW/tbls/compare/v1.73.3...v1.74.0) - 2024-04-13
### Other Changes
- chore(deps): bump the dependencies group with 5 updates by @dependabot in https://github.com/k1LoW/tbls/pull/568
- Update go-github-client version and and use new options in `NewGithubClient`. by @kromiii in https://github.com/k1LoW/tbls/pull/572
- chore(deps): bump the dependencies group with 15 updates by @dependabot in https://github.com/k1LoW/tbls/pull/573

## [v1.73.3](https://github.com/k1LoW/tbls/compare/v1.73.2...v1.73.3) - 2024-03-13
### Other Changes
- Fix some minor typos in README by @mmizutani in https://github.com/k1LoW/tbls/pull/563
- Bump google.golang.org/protobuf from 1.32.0 to 1.33.0 by @dependabot in https://github.com/k1LoW/tbls/pull/565

## [v1.73.2](https://github.com/k1LoW/tbls/compare/v1.72.2...v1.73.2) - 2024-01-25
### Other Changes
- Fix CD pipeline by @k1LoW in https://github.com/k1LoW/tbls/pull/554
- Update pkgs by @k1LoW in https://github.com/k1LoW/tbls/pull/556

## [v1.72.2](https://github.com/k1LoW/tbls/compare/v1.72.1...v1.72.2) - 2024-01-25
### Other Changes
- Update go-graphviz to v0.1.2 by @k1LoW in https://github.com/k1LoW/tbls/pull/552
- Use octocov-action@v1 by @k1LoW in https://github.com/k1LoW/tbls/pull/553

## [v1.72.1](https://github.com/k1LoW/tbls/compare/v1.72.0...v1.72.1) - 2024-01-09
### Other Changes
- chore: unnecessary use of fmt.Sprintf by @testwill in https://github.com/k1LoW/tbls/pull/544
- Add more build environment test by @k1LoW in https://github.com/k1LoW/tbls/pull/546
- Bump golang.org/x/crypto from 0.14.0 to 0.17.0 by @dependabot in https://github.com/k1LoW/tbls/pull/547
- Bump github.com/cloudflare/circl from 1.3.3 to 1.3.7 by @dependabot in https://github.com/k1LoW/tbls/pull/549
- Bump github.com/dvsekhvalnov/jose2go from 1.5.0 to 1.6.0 by @dependabot in https://github.com/k1LoW/tbls/pull/550

## [v1.72.0](https://github.com/k1LoW/tbls/compare/v1.71.1...v1.72.0) - 2023-11-23
### New Features 🎉
- feat: [MongoDB] Support multiple type field by @mrtc0 in https://github.com/k1LoW/tbls/pull/540

## [v1.71.1](https://github.com/k1LoW/tbls/compare/v1.71.0...v1.71.1) - 2023-11-07
### Fix bug 🐛
- fix #535 by @majimaccho in https://github.com/k1LoW/tbls/pull/536
- Fix handling cardinality by @k1LoW in https://github.com/k1LoW/tbls/pull/538

## [v1.71.0](https://github.com/k1LoW/tbls/compare/v1.70.2...v1.71.0) - 2023-10-27
### Breaking Changes 🛠
- feat:refer viewpoints from tables by @macoto1995 in https://github.com/k1LoW/tbls/pull/532
### Other Changes
- Bump google.golang.org/grpc from 1.58.2 to 1.58.3 by @dependabot in https://github.com/k1LoW/tbls/pull/531

## [v1.70.2](https://github.com/k1LoW/tbls/compare/v1.70.1...v1.70.2) - 2023-10-17
### Other Changes
- Add arguments as second sort key for functions. by @corydoraspanda in https://github.com/k1LoW/tbls/pull/527

## [v1.70.1](https://github.com/k1LoW/tbls/compare/v1.70.0...v1.70.1) - 2023-10-12
### Other Changes
- docs: add the installation guide with aqua by @suzuki-shunsuke in https://github.com/k1LoW/tbls/pull/522
- Bump golang.org/x/net from 0.16.0 to 0.17.0 by @dependabot in https://github.com/k1LoW/tbls/pull/524

## [v1.70.0](https://github.com/k1LoW/tbls/compare/v1.69.2...v1.70.0) - 2023-10-08
### New Features 🎉
- feat:requireViewpoints by @macoto1995 in https://github.com/k1LoW/tbls/pull/517
- `coverage` command support loading schema.json by @k1LoW in https://github.com/k1LoW/tbls/pull/518
### Other Changes
- Bump up go and pkg version by @k1LoW in https://github.com/k1LoW/tbls/pull/520

## [v1.69.2](https://github.com/k1LoW/tbls/compare/v1.69.1...v1.69.2) - 2023-10-03
### Other Changes
- Fix the command line argument for goreleaser release command by @mmizutani in https://github.com/k1LoW/tbls/pull/514

## [v1.69.1](https://github.com/k1LoW/tbls/compare/v1.69.0...v1.69.1) - 2023-10-02
### Other Changes
- Fix linux arm64 release build by @mmizutani in https://github.com/k1LoW/tbls/pull/512

## [v1.69.0](https://github.com/k1LoW/tbls/compare/v1.68.2...v1.69.0) - 2023-10-02
### New Features 🎉
- Add an arm64 architecture variant of the Linux build by @mmizutani in https://github.com/k1LoW/tbls/pull/510
### Other Changes
- Added description about Viewpoints in README.md by @macoto1995 in https://github.com/k1LoW/tbls/pull/507
- Add release test by @k1LoW in https://github.com/k1LoW/tbls/pull/511

## [v1.68.2](https://github.com/k1LoW/tbls/compare/v1.68.1...v1.68.2) - 2023-07-26
### Fix bug 🐛
- Add escape double quotes for mermaid by @kaitosawada in https://github.com/k1LoW/tbls/pull/504
### Other Changes
- #478: replace STRING_AGG to allow older MSSQL versions by @Lytchev in https://github.com/k1LoW/tbls/pull/502

## [v1.68.1](https://github.com/k1LoW/tbls/compare/v1.68.0...v1.68.1) - 2023-07-08
### Other Changes
- Bump google.golang.org/grpc from 1.51.0 to 1.53.0 by @dependabot in https://github.com/k1LoW/tbls/pull/500

## [v1.68.0](https://github.com/k1LoW/tbls/compare/v1.67.1...v1.68.0) - 2023-06-23
### Breaking Changes 🛠
- Use schema.json for subcommand by @k1LoW in https://github.com/k1LoW/tbls/pull/497

## [v1.67.1](https://github.com/k1LoW/tbls/compare/v1.67.0...v1.67.1) - 2023-06-18
### New Features 🎉
- Support `--dsn` option for external subcommands by @k1LoW in https://github.com/k1LoW/tbls/pull/495

## [v1.67.0](https://github.com/k1LoW/tbls/compare/v1.66.0...v1.67.0) - 2023-06-17
### New Features 🎉
- Support grouping tables in viewpoints by @k1LoW in https://github.com/k1LoW/tbls/pull/493
### Fix bug 🐛
- Fix schema.Schema.NormalizeTableName by @k1LoW in https://github.com/k1LoW/tbls/pull/490
### Other Changes
- Use CloneWithoutViewpoints in config.Config.ModifySchema by @k1LoW in https://github.com/k1LoW/tbls/pull/492
- Use lo by @k1LoW in https://github.com/k1LoW/tbls/pull/494

## [v1.66.0](https://github.com/k1LoW/tbls/compare/v1.65.4...v1.66.0) - 2023-06-16
### Breaking Changes 🛠
- Sort functions when config.Config.Format.Sort = true by @k1LoW in https://github.com/k1LoW/tbls/pull/484
- Read from config to generate ER diagram or not by @k1LoW in https://github.com/k1LoW/tbls/pull/485
- Support filtering by column labels by @k1LoW in https://github.com/k1LoW/tbls/pull/487
### New Features 🎉
- Support for making viewpoints by @k1LoW in https://github.com/k1LoW/tbls/pull/486
### Fix bug 🐛
- Fix parsing foreign key constraints by @k1LoW in https://github.com/k1LoW/tbls/pull/489
### Other Changes
- Add pronunciation by @muno92 in https://github.com/k1LoW/tbls/pull/481
- Bump github.com/snowflakedb/gosnowflake from 1.6.16 to 1.6.19 by @dependabot in https://github.com/k1LoW/tbls/pull/483
- Add viewpoint validation by @k1LoW in https://github.com/k1LoW/tbls/pull/488

## [v1.65.4](https://github.com/k1LoW/tbls/compare/v1.65.3...v1.65.4) - 2023-06-07
- Fix for when Postgres search path is empty by @codetheweb in https://github.com/k1LoW/tbls/pull/479

## [v1.65.3](https://github.com/k1LoW/tbls/compare/v1.65.2...v1.65.3) - 2023-04-13
- docs: add MacPorts install info by @herbygillot in https://github.com/k1LoW/tbls/pull/473
- Add `--dsn` option by @k1LoW in https://github.com/k1LoW/tbls/pull/475

## [v1.65.2](https://github.com/k1LoW/tbls/compare/v1.65.1...v1.65.2) - 2023-04-10
- Support wildcards ( `*` ) in `--label` option by @k1LoW in https://github.com/k1LoW/tbls/pull/470
- Use schema.json in `tbls out` by @k1LoW in https://github.com/k1LoW/tbls/pull/472

## [v1.65.1](https://github.com/k1LoW/tbls/compare/v1.65.0...v1.65.1) - 2023-04-08
- Fix CI by @k1LoW in https://github.com/k1LoW/tbls/pull/468

## [v1.65.0](https://github.com/k1LoW/tbls/compare/v1.64.1...v1.65.0) - 2023-04-08
- Add lint rule `lint.requireTableLabels:` by @k1LoW in https://github.com/k1LoW/tbls/pull/463
- Add a NamingStrategy to map Post.user_id → User.user_id style detected relation by @yu-ichiro in https://github.com/k1LoW/tbls/pull/466
- Add command `tbls ls` by @k1LoW in https://github.com/k1LoW/tbls/pull/467

## [v1.64.1](https://github.com/k1LoW/tbls/compare/v1.64.0...v1.64.1) - 2023-04-03
- Add more workaround by @k1LoW in https://github.com/k1LoW/tbls/pull/461

## [v1.64.0](https://github.com/k1LoW/tbls/compare/v1.63.0...v1.64.0) - 2023-03-25
- [Breaking Change] Add workaround for overlapping characters in the table name area of ER diagrams. by @k1LoW in https://github.com/k1LoW/tbls/pull/456
- [Breaking Change] Add workaround for overlapping text in table name area of ER diagram on table page by @k1LoW in https://github.com/k1LoW/tbls/pull/458

## [v1.63.0](https://github.com/k1LoW/tbls/compare/v1.62.1...v1.63.0) - 2023-03-13
- Fix unclosed fenced code block in README by @onk in https://github.com/k1LoW/tbls/pull/445
- Add build tag for test by @k1LoW in https://github.com/k1LoW/tbls/pull/451
- Support CHECK constraints for MySQL8 by @k1LoW in https://github.com/k1LoW/tbls/pull/452
- No html encoding is needed in mermaid by @k1LoW in https://github.com/k1LoW/tbls/pull/453

## [v1.62.1](https://github.com/k1LoW/tbls/compare/v1.62.0...v1.62.1) - 2023-02-18
- Bump golang.org/x/image from 0.3.0 to 0.5.0 by @dependabot in https://github.com/k1LoW/tbls/pull/443
- Bump golang.org/x/net from 0.1.0 to 0.7.0 by @dependabot in https://github.com/k1LoW/tbls/pull/444

## [v1.62.0](https://github.com/k1LoW/tbls/compare/v1.61.0...v1.62.0) - 2023-02-08
- Bundle LICENSE file by @k1LoW in https://github.com/k1LoW/tbls/pull/433
- Add `er.showColumnTypes.related:` for showing related columns only by @k1LoW in https://github.com/k1LoW/tbls/pull/432
- Add `er.hideColumnTypes:notRelated:` for hiding not related columns by @yasu89 in https://github.com/k1LoW/tbls/pull/430
- Add `er.showColumnTypes.primary:` for showing primary key columns only by @k1LoW in https://github.com/k1LoW/tbls/pull/435

## [v1.61.0](https://github.com/k1LoW/tbls/compare/v1.60.0...v1.61.0) - 2023-02-02
- Added options for curl to skip downloading of tbls if the newest file exists by @usmanovbf in https://github.com/k1LoW/tbls/pull/423
- Fixed position of a comment in the `use` script by @rsky in https://github.com/k1LoW/tbls/pull/426
- Add test to run installation scripts ( `use` ) by @k1LoW in https://github.com/k1LoW/tbls/pull/427
- Bulk get schema info for MySQL by @yasu89 in https://github.com/k1LoW/tbls/pull/428

## [v1.60.0](https://github.com/k1LoW/tbls/compare/v1.59.0...v1.60.0) - 2023-01-23
- Fix column name prefix for PlantUML by @k1LoW in https://github.com/k1LoW/tbls/pull/418
- Support `github://` by @k1LoW in https://github.com/k1LoW/tbls/pull/419
- Update pkgs by @k1LoW in https://github.com/k1LoW/tbls/pull/421
- [BREAKING] Do not read config file in the default path if env `TBLS_DSN` is given by @k1LoW in https://github.com/k1LoW/tbls/pull/422

## [v1.59.0](https://github.com/k1LoW/tbls/compare/v1.58.0...v1.59.0) - 2023-01-22
- Fix option ( `--table` `--include` `--exclude` ) by @k1LoW in https://github.com/k1LoW/tbls/pull/407
- Support `--label` option for filtering tables by @k1LoW in https://github.com/k1LoW/tbls/pull/409
- Fix test by @k1LoW in https://github.com/k1LoW/tbls/pull/410
- Add `relations[*].override:` section for overriding relations by @k1LoW in https://github.com/k1LoW/tbls/pull/411
- [BREAKING] Support for cardinality detection by @k1LoW in https://github.com/k1LoW/tbls/pull/412
- Support for displaying cardinality in PlantUML output by @k1LoW in https://github.com/k1LoW/tbls/pull/413
- Add Mermaid output format by @k1LoW in https://github.com/k1LoW/tbls/pull/414
- Support `mermaid` for `er.format:` by @k1LoW in https://github.com/k1LoW/tbls/pull/415
- Add testutil by @k1LoW in https://github.com/k1LoW/tbls/pull/416
- Add `er.hideDef:` for hiding relation definition by @k1LoW in https://github.com/k1LoW/tbls/pull/417

## [v1.58.0](https://github.com/k1LoW/tbls/compare/v1.57.1...v1.58.0) - 2023-01-05
- [BREAKING] Exclude all tables except specific table by @k1LoW in https://github.com/k1LoW/tbls/pull/402
- Bump up go and pkgs by @k1LoW in https://github.com/k1LoW/tbls/pull/404
- [BREAKING] Add `--include/--exclude` option / Change logic of `--distance/--table` for tbls out by @k1LoW in https://github.com/k1LoW/tbls/pull/406

## [v1.57.1](https://github.com/k1LoW/tbls/compare/v1.57.0...v1.57.1) - 2022-12-11
- Fix release pipeline (Linux) by @k1LoW in https://github.com/k1LoW/tbls/pull/399

## [v1.57.0](https://github.com/k1LoW/tbls/compare/v1.56.9...v1.57.0) - 2022-12-11
- Fix linter settings by @k1LoW in https://github.com/k1LoW/tbls/pull/394
- Update pkgs by @k1LoW in https://github.com/k1LoW/tbls/pull/396
- Use mongo:4.x because MongoDB 5.0+ requires a CPU with AVX support by @k1LoW in https://github.com/k1LoW/tbls/pull/397
- [BREAKING] Output schema data file by default by @k1LoW in https://github.com/k1LoW/tbls/pull/398

## [v1.56.9](https://github.com/k1LoW/tbls/compare/v1.56.8...v1.56.9) - 2022-12-06
- Update sample/mariadb by @k1LoW in https://github.com/k1LoW/tbls/pull/391
- fix: Percentage columns are always output in Excel format by @zonbitamago in https://github.com/k1LoW/tbls/pull/390
- Change columns key to constants. by @k1LoW in https://github.com/k1LoW/tbls/pull/393

## [v1.56.8](https://github.com/k1LoW/tbls/compare/v1.56.7...v1.56.8) - 2022-11-09
- Support PostgreSQL15 by @k1LoW in https://github.com/k1LoW/tbls/pull/385

## [v1.56.7](https://github.com/k1LoW/tbls/compare/v1.56.6...v1.56.7) - 2022-11-08
- Cross schema foreign keys cannot be procesed by @ypyl in https://github.com/k1LoW/tbls/pull/383

## [v1.56.6](https://github.com/k1LoW/tbls/compare/v1.56.5...v1.56.6) - 2022-10-31
- Copy ca-certificates.crt from builder image by @k1LoW in https://github.com/k1LoW/tbls/pull/379

## [v1.56.5](https://github.com/k1LoW/tbls/compare/v1.56.4...v1.56.5) - 2022-10-26
- Fix docker image build pipeline by @k1LoW in https://github.com/k1LoW/tbls/pull/375
- `tbls lint` line breaks were missing, so added by @k1LoW in https://github.com/k1LoW/tbls/pull/377

## [v1.56.4](https://github.com/k1LoW/tbls/compare/v1.56.3...v1.56.4) - 2022-10-25
- Use tagpr by @k1LoW in https://github.com/k1LoW/tbls/pull/373

## [v1.56.3](https://github.com/k1LoW/tbls/compare/v1.56.2...v1.56.3) (2022-09-24)

* Skip getting functions because Amazon RedShift does not support `pg_get_funciton_arguments()` [#371](https://github.com/k1LoW/tbls/pull/371) ([k1LoW](https://github.com/k1LoW))

## [v1.56.2](https://github.com/k1LoW/tbls/compare/v1.56.1...v1.56.2) (2022-09-02)

* Added sorting to the Table list. [#364](https://github.com/k1LoW/tbls/pull/364) ([awhitford](https://github.com/awhitford))
* Fix extra `-e` displayed on temporary installation. [#365](https://github.com/k1LoW/tbls/pull/365) ([ogumaru](https://github.com/ogumaru))

## [v1.56.1](https://github.com/k1LoW/tbls/compare/v1.56.0...v1.56.1) (2022-07-28)

* Support service account impersonation with BigQuery [#363](https://github.com/k1LoW/tbls/pull/363) ([k1LoW](https://github.com/k1LoW))
* [FEATURE REQUEST] Support service account impersonation with Cloud Spanner [#360](https://github.com/k1LoW/tbls/pull/360) ([Attsun1031](https://github.com/Attsun1031))
* Workaround: change windows platform [#361](https://github.com/k1LoW/tbls/pull/361) ([k1LoW](https://github.com/k1LoW))
* Introduce golangci-lint [#356](https://github.com/k1LoW/tbls/pull/356) ([k1LoW](https://github.com/k1LoW))
* Fix CI / Update go version [#355](https://github.com/k1LoW/tbls/pull/355) ([k1LoW](https://github.com/k1LoW))

## [v1.56.0](https://github.com/k1LoW/tbls/compare/v1.55.1...v1.56.0) (2022-05-28)

* Fix release pipeline [#349](https://github.com/k1LoW/tbls/pull/349) ([k1LoW](https://github.com/k1LoW))
* [BREAKING] Add `format.hideColumnsWithoutValues:` [#354](https://github.com/k1LoW/tbls/pull/354) ([k1LoW](https://github.com/k1LoW))
* Update README [#353](https://github.com/k1LoW/tbls/pull/353) ([bugcloud](https://github.com/bugcloud))
* [BREAKING] Stored procedure/functions support for MsSQL, MySQL and PostrgeSQL [#341](https://github.com/k1LoW/tbls/pull/341) ([YauhenPylAurea](https://github.com/YauhenPylAurea))
* Git.io deprecation [#352](https://github.com/k1LoW/tbls/pull/352) ([k1LoW](https://github.com/k1LoW))
* doc: Add explanation for hide_auto_increment option [#350](https://github.com/k1LoW/tbls/pull/350) ([tk0miya](https://github.com/tk0miya))
* Er diagram links [#347](https://github.com/k1LoW/tbls/pull/347) ([YauhenPylAurea](https://github.com/YauhenPylAurea))

## [v1.55.1](https://github.com/k1LoW/tbls/compare/v1.55.0...v1.55.1) (2022-04-05)

* Revert build env [#345](https://github.com/k1LoW/tbls/pull/345) ([k1LoW](https://github.com/k1LoW))

## [v1.55.0](https://github.com/k1LoW/tbls/compare/v1.54.2...v1.55.0) (2022-04-04)

* Fix release flow [#343](https://github.com/k1LoW/tbls/pull/343) ([k1LoW](https://github.com/k1LoW))
* Mongo basic support [#339](https://github.com/k1LoW/tbls/pull/339) ([YauhenPylAurea](https://github.com/YauhenPylAurea))
* Mdlink encoding [#340](https://github.com/k1LoW/tbls/pull/340) ([YauhenPylAurea](https://github.com/YauhenPylAurea))
* Use octocov [#333](https://github.com/k1LoW/tbls/pull/333) ([k1LoW](https://github.com/k1LoW))
* Fix typo on go install command [#332](https://github.com/k1LoW/tbls/pull/332) ([HMasataka](https://github.com/HMasataka))

## [v1.54.2](https://github.com/k1LoW/tbls/compare/v1.54.1...v1.54.2) (2021-12-16)

* Fix the order of table columns including columnLabels [#331](https://github.com/k1LoW/tbls/pull/331) ([kymmt90](https://github.com/kymmt90))

## [v1.54.1](https://github.com/k1LoW/tbls/compare/v1.54.0...v1.54.1) (2021-12-15)

* Update pkgs [#330](https://github.com/k1LoW/tbls/pull/330) ([k1LoW](https://github.com/k1LoW))
* Update available commands for tbls [#329](https://github.com/k1LoW/tbls/pull/329) ([omuomugin](https://github.com/omuomugin))

## [v1.54.0](https://github.com/k1LoW/tbls/compare/v1.53.0...v1.54.0) (2021-12-10)

* [BREAKING] If it has labels, show the labels in the markdown table [#328](https://github.com/k1LoW/tbls/pull/328) ([k1LoW](https://github.com/k1LoW))
* Add `comments.ColumnLabels:` [#327](https://github.com/k1LoW/tbls/pull/327) ([k1LoW](https://github.com/k1LoW))

## [v1.53.0](https://github.com/k1LoW/tbls/compare/v1.52.0...v1.53.0) (2021-11-16)

* Update packages and Go [#326](https://github.com/k1LoW/tbls/pull/326) ([k1LoW](https://github.com/k1LoW))
* Add `format.showOnlyFirstParagraph:` [#325](https://github.com/k1LoW/tbls/pull/325) ([k1LoW](https://github.com/k1LoW))

## [v1.52.0](https://github.com/k1LoW/tbls/compare/v1.51.0...v1.52.0) (2021-10-22)

* [BREAKING] Use github.com/k1LoW/expand to expand environment variables [#324](https://github.com/k1LoW/tbls/pull/324) ([k1LoW](https://github.com/k1LoW))
* Update Dockerfile [#323](https://github.com/k1LoW/tbls/pull/323) ([k1LoW](https://github.com/k1LoW))
* Use "ghcr.io/k1low/tbls" instead of "k1low/tbls" [#322](https://github.com/k1LoW/tbls/pull/322) ([suzuki](https://github.com/suzuki))

## [v1.51.0](https://github.com/k1LoW/tbls/compare/v1.50.0...v1.51.0) (2021-10-19)

* Add new auto detection strategy for singular table name [#320](https://github.com/k1LoW/tbls/pull/320) ([suzuki](https://github.com/suzuki))
* Replace packr2 to embed [#318](https://github.com/k1LoW/tbls/pull/318) ([k1LoW](https://github.com/k1LoW))
* Support GOOGLE_APPLICATION_CREDENTIALS_JSON [#317](https://github.com/k1LoW/tbls/pull/317) ([k1LoW](https://github.com/k1LoW))
* Replace io/ioutil [#316](https://github.com/k1LoW/tbls/pull/316) ([k1LoW](https://github.com/k1LoW))
* Support darwin arm64 [#315](https://github.com/k1LoW/tbls/pull/315) ([k1LoW](https://github.com/k1LoW))
* Fix pkg vulnerability [#313](https://github.com/k1LoW/tbls/pull/313) ([k1LoW](https://github.com/k1LoW))
* Add config `format.number:` for display sequential numbers in table rows [#312](https://github.com/k1LoW/tbls/pull/312) ([k1LoW](https://github.com/k1LoW))
* Add PostgreSQL SSL mode in README.md [#309](https://github.com/k1LoW/tbls/pull/309) ([kakisoft](https://github.com/kakisoft))
* Bump up go and pkg version [#308](https://github.com/k1LoW/tbls/pull/308) ([k1LoW](https://github.com/k1LoW))
* [BREAKING]Add a list of tables referenced by the view table. [#302](https://github.com/k1LoW/tbls/pull/302) ([k1LoW](https://github.com/k1LoW))

## [v1.50.0](https://github.com/k1LoW/tbls/compare/v1.49.7...v1.50.0) (2021-04-01)

* Add requiredVersion to define a version constraint [#303](https://github.com/k1LoW/tbls/pull/303) ([k1LoW](https://github.com/k1LoW))

## [v1.49.7](https://github.com/k1LoW/tbls/compare/v1.49.6...v1.49.7) (2021-03-06)

* Add Dockerfile for ghcr.io [#301](https://github.com/k1LoW/tbls/pull/301) ([k1LoW](https://github.com/k1LoW))

## [v1.49.6](https://github.com/k1LoW/tbls/compare/v1.49.5...v1.49.6) (2021-02-18)

* Show diff if no target directory [#299](https://github.com/k1LoW/tbls/pull/299) ([k1LoW](https://github.com/k1LoW))
* [Cloud Spanner] Fix order of index columns [#298](https://github.com/k1LoW/tbls/pull/298) ([naoina](https://github.com/naoina))

## [v1.49.5](https://github.com/k1LoW/tbls/compare/v1.49.4...v1.49.5) (2021-02-17)

* Fix 'no such file or directory error' when --rm-dist and no directory [#297](https://github.com/k1LoW/tbls/pull/297) ([k1LoW](https://github.com/k1LoW))
* Minor typos and fixes to issue templates [#296](https://github.com/k1LoW/tbls/pull/296) ([daltonfury42](https://github.com/daltonfury42))

## [v1.49.4](https://github.com/k1LoW/tbls/compare/v1.49.3...v1.49.4) (2021-02-05)

* [MariaDB]Fix bug when table has same name constraints. [#294](https://github.com/k1LoW/tbls/pull/294) ([k1LoW](https://github.com/k1LoW))
* Handling empty env vars in strict mode [#293](https://github.com/k1LoW/tbls/pull/293) ([navaneeth-spotnana](https://github.com/navaneeth-spotnana))
* Add section about updating the documentation [#287](https://github.com/k1LoW/tbls/pull/287) ([tomi](https://github.com/tomi))

## [v1.49.3](https://github.com/k1LoW/tbls/compare/v1.49.2...v1.49.3) (2021-02-04)

* Use RunE [#292](https://github.com/k1LoW/tbls/pull/292) ([k1LoW](https://github.com/k1LoW))
* Fix errors handling [#291](https://github.com/k1LoW/tbls/pull/291) ([k1LoW](https://github.com/k1LoW))

## [v1.49.2](https://github.com/k1LoW/tbls/compare/v1.49.1...v1.49.2) (2021-02-01)

* Add --rm-dist option to remove files in docPath before generating documents [#285](https://github.com/k1LoW/tbls/pull/285) ([k1LoW](https://github.com/k1LoW))
* Fix dsn url (tbls diff output) [#284](https://github.com/k1LoW/tbls/pull/284) ([k1LoW](https://github.com/k1LoW))

## [v1.49.1](https://github.com/k1LoW/tbls/compare/v1.49.0...v1.49.1) (2021-02-01)

* Fix tbls diff (no args) [#283](https://github.com/k1LoW/tbls/pull/283) ([k1LoW](https://github.com/k1LoW))

## [v1.49.0](https://github.com/k1LoW/tbls/compare/v1.48.1...v1.49.0) (2021-02-01)

* Separate datasource/datasource.go [#282](https://github.com/k1LoW/tbls/pull/282) ([k1LoW](https://github.com/k1LoW))
* Fix tbls diff output [#281](https://github.com/k1LoW/tbls/pull/281) ([k1LoW](https://github.com/k1LoW))
* `tbls diff` support for diff checking between dsn and dsn [#280](https://github.com/k1LoW/tbls/pull/280) ([k1LoW](https://github.com/k1LoW))

## [v1.48.1](https://github.com/k1LoW/tbls/compare/v1.48.0...v1.48.1) (2021-01-21)

* [PostgreSQL]Fix version parse [#279](https://github.com/k1LoW/tbls/pull/279) ([k1LoW](https://github.com/k1LoW))

## [v1.48.0](https://github.com/k1LoW/tbls/compare/v1.47.0...v1.48.0) (2021-01-16)

* Add hide_auto_increment option to hide the entire AUTO_INCREMENT clause [#277](https://github.com/k1LoW/tbls/pull/277) ([k1LoW](https://github.com/k1LoW))
* Support MariaDB [#276](https://github.com/k1LoW/tbls/pull/276) ([k1LoW](https://github.com/k1LoW))

## [v1.47.0](https://github.com/k1LoW/tbls/compare/v1.46.0...v1.47.0) (2020-12-31)

* Fix MySQL extra definition [#272](https://github.com/k1LoW/tbls/pull/272) ([k1LoW](https://github.com/k1LoW))
* [BREAKING][PostgreSQL]Support Extra Definition (Generated column) [#271](https://github.com/k1LoW/tbls/pull/271) ([k1LoW](https://github.com/k1LoW))
* Organize testdata [#270](https://github.com/k1LoW/tbls/pull/270) ([k1LoW](https://github.com/k1LoW))
* xlsx format support extra definition [#269](https://github.com/k1LoW/tbls/pull/269) ([k1LoW](https://github.com/k1LoW))
* [BREAKING][MySQL]Support Extra Definition (AUTO_INCREMENT / Generated  column etc) [#268](https://github.com/k1LoW/tbls/pull/268) ([k1LoW](https://github.com/k1LoW))

## [v1.46.0](https://github.com/k1LoW/tbls/compare/v1.45.2...v1.46.0) (2020-12-11)

* Support Snowflake [#267](https://github.com/k1LoW/tbls/pull/267) ([kanata2](https://github.com/kanata2))

## [v1.45.2](https://github.com/k1LoW/tbls/compare/v1.45.1...v1.45.2) (2020-11-26)

* Fix external sub-command logic [#265](https://github.com/k1LoW/tbls/pull/265) ([k1LoW](https://github.com/k1LoW))

## [v1.45.1](https://github.com/k1LoW/tbls/compare/v1.45.0...v1.45.1) (2020-11-23)

* Fix detectVirtualRelations: default strategy [#264](https://github.com/k1LoW/tbls/pull/264) ([k1LoW](https://github.com/k1LoW))

## [v1.45.0](https://github.com/k1LoW/tbls/compare/v1.44.0...v1.45.0) (2020-11-10)

* Add flag that it should deltect relations  [#260](https://github.com/k1LoW/tbls/pull/260) ([syarig](https://github.com/syarig))
* [BigQuery] Sort labels [#263](https://github.com/k1LoW/tbls/pull/263) ([k1LoW](https://github.com/k1LoW))
* Add baseUrl flag for links [#261](https://github.com/k1LoW/tbls/pull/261) ([wubin1989](https://github.com/wubin1989))

## [v1.44.0](https://github.com/k1LoW/tbls/compare/v1.43.1...v1.44.0) (2020-11-05)

* Bump up go version [#262](https://github.com/k1LoW/tbls/pull/262) ([k1LoW](https://github.com/k1LoW))
* Support personalized templates [#258](https://github.com/k1LoW/tbls/pull/258) ([folago](https://github.com/folago))

## [v1.43.1](https://github.com/k1LoW/tbls/compare/v1.43.0...v1.43.1) (2020-08-13)

* Remove array_remove() for supporting PostgreSQL 9.2 [#253](https://github.com/k1LoW/tbls/pull/253) ([k1LoW](https://github.com/k1LoW))
* Fix raws.Close() position [#252](https://github.com/k1LoW/tbls/pull/252) ([k1LoW](https://github.com/k1LoW))

## [v1.43.0](https://github.com/k1LoW/tbls/compare/v1.42.0...v1.43.0) (2020-08-07)

* Add `allOrNothing:` lint option [#250](https://github.com/k1LoW/tbls/pull/250) ([k1LoW](https://github.com/k1LoW))

## [v1.42.0](https://github.com/k1LoW/tbls/compare/v1.41.0...v1.42.0) (2020-08-03)

* [BREAKING] Remove `--add` option [#249](https://github.com/k1LoW/tbls/pull/249) ([k1LoW](https://github.com/k1LoW))
* Add --when option [#248](https://github.com/k1LoW/tbls/pull/248) ([k1LoW](https://github.com/k1LoW))
* [BREAKING] Move `tbls out -t cacoo` to `tbls cacoo csv` (tbls-cacoo) [#247](https://github.com/k1LoW/tbls/pull/247) ([k1LoW](https://github.com/k1LoW))

## [v1.40.0](https://github.com/k1LoW/tbls/compare/v1.39.0...v1.40.0) (2020-07-09)

* [BREAKING] Change default ER diagram format to 'svg' [#244](https://github.com/k1LoW/tbls/pull/244) ([k1LoW](https://github.com/k1LoW))
* Add er.font option for png/jpg ER diagram [#243](https://github.com/k1LoW/tbls/pull/243) ([k1LoW](https://github.com/k1LoW))

## [v1.39.0](https://github.com/k1LoW/tbls/compare/v1.38.3...v1.39.0) (2020-06-15)

* [BREAKING][MySQL]Redact AUTO_INCREMENT value of table option by default [#241](https://github.com/k1LoW/tbls/pull/241) ([k1LoW](https://github.com/k1LoW))

## [v1.38.3](https://github.com/k1LoW/tbls/compare/v1.38.2...v1.38.3) (2020-05-13)

* Fix MSSQL comments trimmed to 30 chars [#238](https://github.com/k1LoW/tbls/pull/238) ([paulKabira](https://github.com/paulKabira))

## [v1.38.2](https://github.com/k1LoW/tbls/compare/v1.38.1...v1.38.2) (2020-05-11)

* [BREAKING]Set path of temporary generated schema.json (instead of JSON string) to TBLS_SCHEMA [#236](https://github.com/k1LoW/tbls/pull/236) ([k1LoW](https://github.com/k1LoW))

## [v1.38.1](https://github.com/k1LoW/tbls/compare/v1.38.0...v1.38.1) (2020-05-10)

* Fix Dockerfile.hub.docker.com [#235](https://github.com/k1LoW/tbls/pull/235) ([k1LoW](https://github.com/k1LoW))

## [v1.38.0](https://github.com/k1LoW/tbls/compare/v1.37.5...v1.38.0) (2020-05-10)

* Support MSSQL Description for Table/Columns. [#234](https://github.com/k1LoW/tbls/pull/234) ([paulKabira](https://github.com/paulKabira))

## [v1.37.5](https://github.com/k1LoW/tbls/compare/v1.37.4...v1.37.5) (2020-05-09)

* make config.DefaultDocPath and config.DefaultConfigFilePaths public [#233](https://github.com/k1LoW/tbls/pull/233) ([k1LoW](https://github.com/k1LoW))
* fix typo [#231](https://github.com/k1LoW/tbls/pull/231) ([dojineko](https://github.com/dojineko))

## [v1.37.4](https://github.com/k1LoW/tbls/compare/v1.37.3...v1.37.4) (2020-05-07)

* Pass env `TBLS_CONFIG_PATH` to subcommands [#230](https://github.com/k1LoW/tbls/pull/230) ([k1LoW](https://github.com/k1LoW))

## [v1.37.3](https://github.com/k1LoW/tbls/compare/v1.37.2...v1.37.3) (2020-05-05)

* Support completion of external subcommands (Bash) [#229](https://github.com/k1LoW/tbls/pull/229) ([k1LoW](https://github.com/k1LoW))
* Fix basename issue and simplify code [#228](https://github.com/k1LoW/tbls/pull/228) ([syohex](https://github.com/syohex))
* Support completion of external subcommands (Zsh) [#227](https://github.com/k1LoW/tbls/pull/227) ([k1LoW](https://github.com/k1LoW))

## [v1.37.2](https://github.com/k1LoW/tbls/compare/v1.37.1...v1.37.2) (2020-05-01)

* Remove relations of excluded tables [#226](https://github.com/k1LoW/tbls/pull/226) ([k1LoW](https://github.com/k1LoW))

## [v1.37.1](https://github.com/k1LoW/tbls/compare/v1.37.0...v1.37.1) (2020-04-30)

* Fix FilterTables [#225](https://github.com/k1LoW/tbls/pull/225) ([k1LoW](https://github.com/k1LoW))
* Fix usage [#224](https://github.com/k1LoW/tbls/pull/224) ([k1LoW](https://github.com/k1LoW))

## [v1.37.0](https://github.com/k1LoW/tbls/compare/v1.36.1...v1.37.0) (2020-04-28)

* Add `tbls coverage` [#222](https://github.com/k1LoW/tbls/pull/222) ([k1LoW](https://github.com/k1LoW))
* Add lint rule for comments ( index, constraint, trigger ) [#221](https://github.com/k1LoW/tbls/pull/221) ([k1LoW](https://github.com/k1LoW))
* [PostgreSQL]Support comments ( index, constraint, trigger ) [#220](https://github.com/k1LoW/tbls/pull/220) ([k1LoW](https://github.com/k1LoW))
* Support more comments ( index, constraint, trigger ) [#219](https://github.com/k1LoW/tbls/pull/219) ([k1LoW](https://github.com/k1LoW))

## [v1.36.1](https://github.com/k1LoW/tbls/compare/v1.36.0...v1.36.1) (2020-04-25)

* Add ",omitempty" to config.Config [#218](https://github.com/k1LoW/tbls/pull/218) ([k1LoW](https://github.com/k1LoW))
* Correct test for varchar arrays [#217](https://github.com/k1LoW/tbls/pull/217) ([mjpieters](https://github.com/mjpieters))

## [v1.36.0](https://github.com/k1LoW/tbls/compare/v1.35.0...v1.36.0) (2020-04-25)

* [PostgreSQL]Fix the logic of extracting table/column name from definition of FK [#215](https://github.com/k1LoW/tbls/pull/215) ([k1LoW](https://github.com/k1LoW))
* Postgres: support materialized views [#214](https://github.com/k1LoW/tbls/pull/214) ([mjpieters](https://github.com/mjpieters))

## [v1.35.0](https://github.com/k1LoW/tbls/compare/v1.34.1...v1.35.0) (2020-04-22)

* [BREAKING]Fix `tbls diff` output [#207](https://github.com/k1LoW/tbls/pull/207) ([k1LoW](https://github.com/k1LoW))
* Add datasource.NewBigqueryClient [#206](https://github.com/k1LoW/tbls/pull/206) ([k1LoW](https://github.com/k1LoW))

## [v1.34.1](https://github.com/k1LoW/tbls/compare/v1.34.0...v1.34.1) (2020-04-21)

* Add datasource.AnalyzeJSONString for external subcommand [#205](https://github.com/k1LoW/tbls/pull/205) ([k1LoW](https://github.com/k1LoW))

## [v1.34.0](https://github.com/k1LoW/tbls/compare/v1.33.0...v1.34.0) (2020-04-21)

* Support external subcommand like `git-*` [#204](https://github.com/k1LoW/tbls/pull/204) ([k1LoW](https://github.com/k1LoW))
* Fix typo [#203](https://github.com/k1LoW/tbls/pull/203) ([k1LoW](https://github.com/k1LoW))

## [v1.33.0](https://github.com/k1LoW/tbls/compare/v1.32.2...v1.33.0) (2020-04-19)

* Add output format (CSV for Cacoo's Database Schema Importer) [#202](https://github.com/k1LoW/tbls/pull/202) ([k1LoW](https://github.com/k1LoW))

## [v1.32.2](https://github.com/k1LoW/tbls/compare/v1.32.1...v1.32.2) (2020-04-17)

* Fix lint labelStyleBigQuery [#201](https://github.com/k1LoW/tbls/pull/201) ([k1LoW](https://github.com/k1LoW))

## [v1.32.1](https://github.com/k1LoW/tbls/compare/v1.32.0...v1.32.1) (2020-04-17)

* [BREAKING]Fix `excludedTables:` -> `excludeTables:` [#200](https://github.com/k1LoW/tbls/pull/200) ([k1LoW](https://github.com/k1LoW))

## [v1.32.0](https://github.com/k1LoW/tbls/compare/v1.31.2...v1.32.0) (2020-04-17)

* Add lint rule `labelStyleBigQuery` [#199](https://github.com/k1LoW/tbls/pull/199) ([k1LoW](https://github.com/k1LoW))
* Fix Config.MaskedDSN output [#198](https://github.com/k1LoW/tbls/pull/198) ([k1LoW](https://github.com/k1LoW))

## [v1.31.2](https://github.com/k1LoW/tbls/compare/v1.31.1...v1.31.2) (2020-04-14)

* fix (MSSQLDriver) : Set size of nvarchar/varchar/varbinary columns.  Filter sysname columns from list. [#197](https://github.com/k1LoW/tbls/pull/197) ([jafin](https://github.com/jafin))

## [v1.31.1](https://github.com/k1LoW/tbls/compare/v1.31.0...v1.31.1) (2020-04-14)

* Fixed an error in Graphviz when the parent table of Relation is an exclude target. [#196](https://github.com/k1LoW/tbls/pull/196) ([yoskhdia](https://github.com/yoskhdia))

## [v1.31.0](https://github.com/k1LoW/tbls/compare/v1.30.0...v1.31.0) (2020-04-07)

* Add labels: and comments.labels: [#195](https://github.com/k1LoW/tbls/pull/195) ([k1LoW](https://github.com/k1LoW))
* [BigQuery]Add labels section [#194](https://github.com/k1LoW/tbls/pull/194) ([k1LoW](https://github.com/k1LoW))

## [v1.30.0](https://github.com/k1LoW/tbls/compare/v1.29.3...v1.30.0) (2020-04-06)

* Fix #188 'near "." Syntax Error' [#193](https://github.com/k1LoW/tbls/pull/193) ([BoringDude](https://github.com/BoringDude))
* [BREAKING]Add schema description [#192](https://github.com/k1LoW/tbls/pull/192) ([k1LoW](https://github.com/k1LoW))

## [v1.29.3](https://github.com/k1LoW/tbls/compare/v1.29.2...v1.29.3) (2020-04-05)

* Fix os.OpenFile mode and permission [#191](https://github.com/k1LoW/tbls/pull/191) ([k1LoW](https://github.com/k1LoW))
* Set dict for BigQuery [#190](https://github.com/k1LoW/tbls/pull/190) ([k1LoW](https://github.com/k1LoW))
* Add testing platform ( GitBash on Windows ) [#189](https://github.com/k1LoW/tbls/pull/189) ([k1LoW](https://github.com/k1LoW))
* Support `https://` `http://` [#187](https://github.com/k1LoW/tbls/pull/187) ([k1LoW](https://github.com/k1LoW))
* Add default config file name `tbls.yml` [#186](https://github.com/k1LoW/tbls/pull/186) ([k1LoW](https://github.com/k1LoW))

## [v1.29.2](https://github.com/k1LoW/tbls/compare/v1.29.1...v1.29.2) (2020-03-17)

* Fix syntax error (<br> -> <br />) in dot file [#185](https://github.com/k1LoW/tbls/pull/185) ([k1LoW](https://github.com/k1LoW))

## [v1.29.1](https://github.com/k1LoW/tbls/compare/v1.29.0...v1.29.1) (2020-03-15)

* [PostgreSQL]Detect full table name using search_path and information_schema [#184](https://github.com/k1LoW/tbls/pull/184) ([k1LoW](https://github.com/k1LoW))

## [v1.29.0](https://github.com/k1LoW/tbls/compare/v1.28.2...v1.29.0) (2020-03-14)

* Set dict for DynamoDB [#183](https://github.com/k1LoW/tbls/pull/183) ([k1LoW](https://github.com/k1LoW))
* Add `dict:` for replacement title/table header of database document [#182](https://github.com/k1LoW/tbls/pull/182) ([k1LoW](https://github.com/k1LoW))
* Add `name:` for specifing database name of document [#180](https://github.com/k1LoW/tbls/pull/180) ([k1LoW](https://github.com/k1LoW))
* Fix schema.Driver.Meta [#179](https://github.com/k1LoW/tbls/pull/179) ([k1LoW](https://github.com/k1LoW))

## [v1.28.2](https://github.com/k1LoW/tbls/compare/v1.28.1...v1.28.2) (2020-03-09)

* Support table name omitting current_schema [#178](https://github.com/k1LoW/tbls/pull/178) ([k1LoW](https://github.com/k1LoW))

## [v1.28.1](https://github.com/k1LoW/tbls/compare/v1.28.0...v1.28.1) (2020-03-08)

* Use goccy/go-yaml for loading config, too [#177](https://github.com/k1LoW/tbls/pull/177) ([k1LoW](https://github.com/k1LoW))

## [v1.28.0](https://github.com/k1LoW/tbls/compare/v1.27.0...v1.28.0) (2020-03-07)

* Fix zsh completion [#176](https://github.com/k1LoW/tbls/pull/176) ([k1LoW](https://github.com/k1LoW))
* [BREAKING][PostgreSQL][Amazon Redshift]Show `public.` schema [#175](https://github.com/k1LoW/tbls/pull/175) ([k1LoW](https://github.com/k1LoW))
* Support Amazon DynamoDB [#174](https://github.com/k1LoW/tbls/pull/174) ([k1LoW](https://github.com/k1LoW))
* Filter target tables using `include:` and `exclude:` / `include:` and `exclude:` support wildcard [#172](https://github.com/k1LoW/tbls/pull/172) ([k1LoW](https://github.com/k1LoW))

## [v1.27.0](https://github.com/k1LoW/tbls/compare/v1.26.0...v1.27.0) (2020-02-24)

* Add config.ER.distance [#171](https://github.com/k1LoW/tbls/pull/171) ([k1LoW](https://github.com/k1LoW))
* Fix: `tbls out -t config` does not set default values [#170](https://github.com/k1LoW/tbls/pull/170) ([k1LoW](https://github.com/k1LoW))
* Add output format (PNG, SVG, JPG) [#169](https://github.com/k1LoW/tbls/pull/169) ([k1LoW](https://github.com/k1LoW))
* Use github.com/goccy/go-graphviz instead of dot command [#167](https://github.com/k1LoW/tbls/pull/167) ([k1LoW](https://github.com/k1LoW))
* Bump up goccy/go-yaml version to v1.3.2 and remove workaround [#168](https://github.com/k1LoW/tbls/pull/168) ([k1LoW](https://github.com/k1LoW))
* Change default shell to /bin/sh [#166](https://github.com/k1LoW/tbls/pull/166) ([kkznch](https://github.com/kkznch))

## [v1.26.0](https://github.com/k1LoW/tbls/compare/v1.25.1...v1.26.0) (2020-02-20)

* [BREAKING] Normalize `relations:` of tbls output [#165](https://github.com/k1LoW/tbls/pull/165) ([k1LoW](https://github.com/k1LoW))
* Rename `schema.Relation.IsAdditional` -> `schema.Relation.Virtual` [#164](https://github.com/k1LoW/tbls/pull/164) ([k1LoW](https://github.com/k1LoW))
* Add YAML output format [#163](https://github.com/k1LoW/tbls/pull/163) ([k1LoW](https://github.com/k1LoW))

## [v1.25.1](https://github.com/k1LoW/tbls/compare/v1.25.0...v1.25.1) (2020-02-17)

* [MySQL]Fix constraints/indexes detection bug [#162](https://github.com/k1LoW/tbls/pull/162) ([k1LoW](https://github.com/k1LoW))
* Fix: Empty array is output as null [#161](https://github.com/k1LoW/tbls/pull/161) ([k1LoW](https://github.com/k1LoW))

## [v1.25.0](https://github.com/k1LoW/tbls/compare/v1.24.1...v1.25.0) (2020-02-06)

* lint `exclude:` `lintExclude:` support wildcard (`*`) [#160](https://github.com/k1LoW/tbls/pull/160) ([k1LoW](https://github.com/k1LoW))

## [v1.24.1](https://github.com/k1LoW/tbls/compare/v1.24.0...v1.24.1) (2020-01-16)

* [PostgreSQL]Fix parsing double-quoted table/column name in definition of Foreign Key [#154](https://github.com/k1LoW/tbls/pull/154) ([k1LoW](https://github.com/k1LoW))
* Update README.md [#151](https://github.com/k1LoW/tbls/pull/151) ([ednawig](https://github.com/ednawig))
* Dockerfile use latest version of tbls [#150](https://github.com/k1LoW/tbls/pull/150) ([k1LoW](https://github.com/k1LoW))

## [v1.24.0](https://github.com/k1LoW/tbls/compare/v1.23.0...v1.24.0) (2020-01-11)

* Add `tbls completion` [#149](https://github.com/k1LoW/tbls/pull/149) ([k1LoW](https://github.com/k1LoW))

## [v1.23.0](https://github.com/k1LoW/tbls/compare/v1.22.1...v1.23.0) (2019-12-11)

* Show table names when lint error `unrelated table exists` [#148](https://github.com/k1LoW/tbls/pull/148) ([k1LoW](https://github.com/k1LoW))

## [v1.22.1](https://github.com/k1LoW/tbls/compare/v1.22.0...v1.22.1) (2019-11-21)

* Fix: panic: assignment to entry in nil map [#147](https://github.com/k1LoW/tbls/pull/147) ([k1LoW](https://github.com/k1LoW))

## [v1.22.0](https://github.com/k1LoW/tbls/compare/v1.21.1...v1.22.0) (2019-11-20)

* Add .tbls.yml (tbls config file) output format [#146](https://github.com/k1LoW/tbls/pull/146) ([k1LoW](https://github.com/k1LoW))

## [v1.21.1](https://github.com/k1LoW/tbls/compare/v1.21.0...v1.21.1) (2019-11-11)

* Fix: requireForeignKeyIndex check only FOREIGN_KEY. [#145](https://github.com/k1LoW/tbls/pull/145) ([k1LoW](https://github.com/k1LoW))

## [v1.21.0](https://github.com/k1LoW/tbls/compare/v1.20.0...v1.21.0) (2019-11-11)

* Add `lintExclude` for exclude tables from lint. [#144](https://github.com/k1LoW/tbls/pull/144) ([k1LoW](https://github.com/k1LoW))

## [v1.20.0](https://github.com/k1LoW/tbls/compare/v1.19.0...v1.20.0) (2019-11-02)

* lint rule `requireColumnComment` excludes `table_name.column_name` as well as `column_name` [#143](https://github.com/k1LoW/tbls/pull/143) ([k1LoW](https://github.com/k1LoW))
* Add lint rule `requireForeignKeyIndex` [#142](https://github.com/k1LoW/tbls/pull/142) ([k1LoW](https://github.com/k1LoW))
* Use GitHub Actions [#141](https://github.com/k1LoW/tbls/pull/141) ([k1LoW](https://github.com/k1LoW))
* Remove `MATCH SIMPLE` from my.sql [#140](https://github.com/k1LoW/tbls/pull/140) ([k1LoW](https://github.com/k1LoW))

## [v1.19.0](https://github.com/k1LoW/tbls/compare/v1.18.2...v1.19.0) (2019-09-06)

* Add lint rule `duplicateRelations` [#139](https://github.com/k1LoW/tbls/pull/139) ([k1LoW](https://github.com/k1LoW))

## [v1.18.2](https://github.com/k1LoW/tbls/compare/v1.18.1...v1.18.2) (2019-09-06)

* Remove duplicate relation links [#138](https://github.com/k1LoW/tbls/pull/138) ([k1LoW](https://github.com/k1LoW))
* add error handling [#137](https://github.com/k1LoW/tbls/pull/137) ([toshi0607](https://github.com/toshi0607))
* add Dockerfile [#136](https://github.com/k1LoW/tbls/pull/136) ([peccu](https://github.com/peccu))
* Add gosec [#135](https://github.com/k1LoW/tbls/pull/135) ([k1LoW](https://github.com/k1LoW))

## [v1.18.1](https://github.com/k1LoW/tbls/compare/v1.18.0...v1.18.1) (2019-08-15)

* Fix duplicate output when multiple schemas have the same named table [#134](https://github.com/k1LoW/tbls/pull/134) ([oohira](https://github.com/oohira))
* Support `span://` for Cloud Spanner scheme [#133](https://github.com/k1LoW/tbls/pull/133) ([k1LoW](https://github.com/k1LoW))

## [v1.18.0](https://github.com/k1LoW/tbls/compare/v1.17.2...v1.18.0) (2019-08-13)

* Support Cloud Spanner [#132](https://github.com/k1LoW/tbls/pull/132) ([k1LoW](https://github.com/k1LoW))
* Fix .travis.yml condition [#131](https://github.com/k1LoW/tbls/pull/131) ([k1LoW](https://github.com/k1LoW))
* fix typo for readme [#130](https://github.com/k1LoW/tbls/pull/130) ([kojirock5260](https://github.com/kojirock5260))

## [v1.17.2](https://github.com/k1LoW/tbls/compare/v1.17.1...v1.17.2) (2019-07-15)

* Fix .goreleaser build hooks [#129](https://github.com/k1LoW/tbls/pull/129) ([k1LoW](https://github.com/k1LoW))

## [v1.17.1](https://github.com/k1LoW/tbls/compare/v1.17.0...v1.17.1) (2019-07-08)

* Fix panic when `hyphen-table` [#126](https://github.com/k1LoW/tbls/pull/126) ([k1LoW](https://github.com/k1LoW))
* Update gobuffalo/packr to v2 [#124](https://github.com/k1LoW/tbls/pull/124) ([k1LoW](https://github.com/k1LoW))

## [v1.17.0](https://github.com/k1LoW/tbls/compare/v1.16.1...v1.17.0) (2019-06-12)

* Refactor out/* packages [#123](https://github.com/k1LoW/tbls/pull/123) ([k1LoW](https://github.com/k1LoW))
* Add er.comment for add comment to ER diagram [#122](https://github.com/k1LoW/tbls/pull/122) ([k1LoW](https://github.com/k1LoW))

## [v1.16.1](https://github.com/k1LoW/tbls/compare/v1.16.0...v1.16.1) (2019-06-11)

* Fix loading ER diagram format from .tbls.yml [#120](https://github.com/k1LoW/tbls/pull/120) ([k1LoW](https://github.com/k1LoW))

## [v1.16.0](https://github.com/k1LoW/tbls/compare/v1.15.4...v1.16.0) (2019-06-04)

* Support for Microsoft SQL Server [#118](https://github.com/k1LoW/tbls/pull/118) ([k1LoW](https://github.com/k1LoW))

## [v1.15.4](https://github.com/k1LoW/tbls/compare/v1.15.3...v1.15.4) (2019-05-28)

* Postgres driver support new schema.Constraint schema.Index [#116](https://github.com/k1LoW/tbls/pull/116) ([k1LoW](https://github.com/k1LoW))

## [v1.15.3](https://github.com/k1LoW/tbls/compare/v1.15.2...v1.15.3) (2019-05-27)

* Redshift can not analyze indexes [#115](https://github.com/k1LoW/tbls/pull/115) ([k1LoW](https://github.com/k1LoW))
* Fixed a typo in "repository" [#112](https://github.com/k1LoW/tbls/pull/112) ([AntonNguyen](https://github.com/AntonNguyen))

## [v1.15.2](https://github.com/k1LoW/tbls/compare/v1.15.1...v1.15.2) (2019-05-27)

* Redshift can not analyze constraints [#114](https://github.com/k1LoW/tbls/pull/114) ([k1LoW](https://github.com/k1LoW))
* Revert Postgres driver parsing logic schema.Constaint/schema.Index [#113](https://github.com/k1LoW/tbls/pull/113) ([k1LoW](https://github.com/k1LoW))

## [v1.15.0](https://github.com/k1LoW/tbls/compare/v1.14.0...v1.15.0) (2019-05-26)

* Fix PlantUML output format [#109](https://github.com/k1LoW/tbls/pull/109) ([k1LoW](https://github.com/k1LoW))
* Fix schema.Index [#108](https://github.com/k1LoW/tbls/pull/108) ([k1LoW](https://github.com/k1LoW))
* Fix schema.Constraint [#107](https://github.com/k1LoW/tbls/pull/107) ([k1LoW](https://github.com/k1LoW))
* Add PlantUML output format [#106](https://github.com/k1LoW/tbls/pull/106) ([k1LoW](https://github.com/k1LoW))

## [v1.14.0](https://github.com/k1LoW/tbls/compare/v1.13.3...v1.14.0) (2019-05-15)

* Support Amazon Redshift [#105](https://github.com/k1LoW/tbls/pull/105) ([k1LoW](https://github.com/k1LoW))

## [v1.13.3](https://github.com/k1LoW/tbls/compare/v1.13.2...v1.13.3) (2019-05-13)

* Fix config.RequireColumns is not config.Rule [#104](https://github.com/k1LoW/tbls/pull/104) ([k1LoW](https://github.com/k1LoW))

## [v1.13.2](https://github.com/k1LoW/tbls/compare/v1.13.1...v1.13.2) (2019-05-12)

* Truncate xlsx worksheet name for MS Excel [#102](https://github.com/k1LoW/tbls/pull/102) ([k1LoW](https://github.com/k1LoW))

## [v1.13.1](https://github.com/k1LoW/tbls/compare/v1.13.0...v1.13.1) (2019-05-12)

* Fix BigQuery dataset schema name [#101](https://github.com/k1LoW/tbls/pull/101) ([k1LoW](https://github.com/k1LoW))

## [v1.13.0](https://github.com/k1LoW/tbls/compare/v1.12.0...v1.13.0) (2019-05-12)

* Fix dot file format [#100](https://github.com/k1LoW/tbls/pull/100) ([k1LoW](https://github.com/k1LoW))
* Fix command options [#99](https://github.com/k1LoW/tbls/pull/99) ([k1LoW](https://github.com/k1LoW))
* Support BigQuery [#98](https://github.com/k1LoW/tbls/pull/98) ([k1LoW](https://github.com/k1LoW))
* Refactor drivers [#97](https://github.com/k1LoW/tbls/pull/97) ([k1LoW](https://github.com/k1LoW))

## [v1.12.0](https://github.com/k1LoW/tbls/compare/v1.11.1...v1.12.0) (2019-05-11)

*  Add `exclude` for excluding tables from the document [#96](https://github.com/k1LoW/tbls/pull/96) ([k1LoW](https://github.com/k1LoW))
* Add lint rule `requireColumns` [#95](https://github.com/k1LoW/tbls/pull/95) ([k1LoW](https://github.com/k1LoW))

## [v1.11.1](https://github.com/k1LoW/tbls/compare/v1.11.0...v1.11.1) (2019-04-25)

* Fix loading args when `tbls out` [#91](https://github.com/k1LoW/tbls/pull/91) ([k1LoW](https://github.com/k1LoW))

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
