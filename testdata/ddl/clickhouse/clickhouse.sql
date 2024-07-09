-- https://clickhouse.com/docs/en/sql-reference/statements/create/table
DROP TABLE IF EXISTS testdb.table_name;
CREATE TABLE IF NOT EXISTS testdb.table_name
(
    name1 UInt64 COMMENT 'comment for column 1',
    name2 Nullable(String) DEFAULT 'column 2' COMMENT 'comment for column 2' CODEC (ZSTD),
    name3 LowCardinality(String) MATERIALIZED upper(name2) COMMENT 'comment for column 3',
    name4 SimpleAggregateFunction(sum, DOUBLE) TTL name5 + INTERVAL 1 DAY,
    name5 DateTime         DEFAULT now(),
    name6 String ALIAS formatReadableSize(name1),
    name7 String MATERIALIZED hex(name1),
    name8 FixedString(4)   DEFAULT unhex(name7),

    INDEX idx1 name1 TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx2 name1 * 2 TYPE minmax GRANULARITY 3,
    INDEX idx3 name1 * length(name2) TYPE set(1000) GRANULARITY 4,

    PROJECTION projection_name_1 (SELECT name1, name2, name3
                                  ORDER BY name1)
) ENGINE = MergeTree
      ORDER BY (name1, name5)
      PARTITION BY (name1, name3, name5)
      PRIMARY KEY (name1, name5)
      SAMPLE BY (name1)
      COMMENT 'comment for table';

-- https://clickhouse.com/docs/en/sql-reference/statements/create/table#from-a-table-function
DROP TABLE IF EXISTS testdb.numbers_table;
CREATE TABLE IF NOT EXISTS testdb.numbers_table AS numbers(100);

DROP TABLE IF EXISTS testdb.t1;
CREATE TABLE testdb.t1
(
    x String
) ENGINE = Memory AS
SELECT 1;

-- https://clickhouse.com/docs/en/sql-reference/statements/create/dictionary
DROP TABLE IF EXISTS testdb.source_table;
CREATE TABLE IF NOT EXISTS testdb.source_table
(
    id    UInt64,
    value String
) ENGINE = MergeTree
      PRIMARY KEY id;

DROP DICTIONARY IF EXISTS testdb.id_value_dictionary;
CREATE DICTIONARY testdb.id_value_dictionary
(
    id    UInt64,
    value String
)
    PRIMARY KEY id
    SOURCE (CLICKHOUSE(TABLE 'source_table'))
    LAYOUT (FLAT())
    LIFETIME (MIN 0 MAX 1000);

-- https://clickhouse.com/docs/en/sql-reference/statements/create/function
CREATE FUNCTION linear_equation AS(x, k, b) -> k * x + b;

-- https://clickhouse.com/docs/en/sql-reference/statements/create/view
DROP VIEW IF EXISTS testdb.view;
CREATE VIEW IF NOT EXISTS testdb.view AS
SELECT *
FROM testdb.table_name;

-- https://clickhouse.com/docs/en/sql-reference/statements/create/view#materialized-view
DROP VIEW IF EXISTS testdb.materialized_view;
CREATE MATERIALIZED VIEW IF NOT EXISTS testdb.materialized_view
    ENGINE = Memory
AS
SELECT name1, name2
FROM testdb.table_name
ORDER BY name1 DESC;
