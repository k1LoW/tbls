-- Databricks SQL DDL for TPC-H SF1 benchmark schema
-- Converted from Snowflake DDL with proper constraints
-- ref: https://docs.snowflake.com/en/user-guide/sample-data-tpch.html#database-entities-relationships-and-characteristics

CREATE SCHEMA IF NOT EXISTS TPCH_SF1;
USE SCHEMA TPCH_SF1;

-- Reference tables (no foreign keys)
CREATE TABLE IF NOT EXISTS REGION (
  R_REGIONKEY	BIGINT NOT NULL COMMENT 'Unique identifier for each geographic region',
  R_NAME	STRING NOT NULL COMMENT 'Name of the geographic region (e.g. AMERICA, ASIA)',
  R_COMMENT	STRING COMMENT 'General comments about the region',
  CONSTRAINT region_pk PRIMARY KEY (R_REGIONKEY)
) COMMENT 'Geographic regions containing nations';

CREATE TABLE IF NOT EXISTS NATION (
  N_NATIONKEY	BIGINT NOT NULL COMMENT 'Unique identifier for each nation/country',
  N_NAME	STRING NOT NULL COMMENT 'Name of the nation/country',
  N_REGIONKEY	BIGINT NOT NULL COMMENT 'Foreign key to the region this nation belongs to',
  N_COMMENT	STRING COMMENT 'General comments about the nation',
  CONSTRAINT nation_pk PRIMARY KEY (N_NATIONKEY),
  CONSTRAINT nation_regionkey_fk FOREIGN KEY (N_REGIONKEY) REFERENCES REGION(R_REGIONKEY)
) COMMENT 'Countries/nations with their associated regions';

CREATE TABLE IF NOT EXISTS PART (
  P_PARTKEY	BIGINT NOT NULL COMMENT 'Unique identifier for each part in the catalog',
  P_NAME	STRING NOT NULL COMMENT 'Name of the part',
  P_MFGR	STRING NOT NULL COMMENT 'Manufacturer of the part',
  P_BRAND	STRING NOT NULL COMMENT 'Brand name of the part',
  P_TYPE	STRING NOT NULL COMMENT 'Type/category of the part',
  P_SIZE	BIGINT NOT NULL COMMENT 'Size of the part in standardized units',
  P_CONTAINER	STRING NOT NULL COMMENT 'Container type for shipping the part',
  P_RETAILPRICE	DECIMAL(12,2) NOT NULL COMMENT 'Retail price of the part',
  P_COMMENT	STRING COMMENT 'General comments about the part',
  CONSTRAINT part_pk PRIMARY KEY (P_PARTKEY)
) COMMENT 'Parts catalog with manufacturing and pricing information';

CREATE TABLE IF NOT EXISTS SUPPLIER (
  S_SUPPKEY	BIGINT NOT NULL COMMENT 'Unique identifier for each supplier',
  S_NAME	STRING NOT NULL COMMENT 'Name of the supplier company',
  S_ADDRESS	STRING NOT NULL COMMENT 'Complete address of the supplier',
  S_NATIONKEY	BIGINT NOT NULL COMMENT 'Foreign key to the nation where supplier is located',
  S_PHONE	STRING NOT NULL COMMENT 'Phone number of the supplier',
  S_ACCTBAL	DECIMAL(12,2) NOT NULL COMMENT 'Account balance with this supplier',
  S_COMMENT	STRING COMMENT 'General comments about the supplier',
  CONSTRAINT supplier_pk PRIMARY KEY (S_SUPPKEY),
  CONSTRAINT supplier_nationkey_fk FOREIGN KEY (S_NATIONKEY) REFERENCES NATION(N_NATIONKEY)
) COMMENT 'Suppliers with contact and account information';

CREATE TABLE IF NOT EXISTS CUSTOMER (
  C_CUSTKEY	BIGINT NOT NULL COMMENT 'Unique identifier for each customer',
  C_NAME	STRING NOT NULL COMMENT 'Name of the customer',
  C_ADDRESS	STRING NOT NULL COMMENT 'Complete address of the customer',
  C_NATIONKEY	BIGINT NOT NULL COMMENT 'Foreign key to the nation where customer is located',
  C_PHONE	STRING NOT NULL COMMENT 'Phone number of the customer',
  C_ACCTBAL	DECIMAL(12,2) NOT NULL COMMENT 'Account balance of the customer',
  C_MKTSEGMENT	STRING COMMENT 'Market segment classification (e.g. BUILDING, AUTOMOBILE)',
  C_COMMENT	STRING COMMENT 'General comments about the customer',
  CONSTRAINT customer_pk PRIMARY KEY (C_CUSTKEY),
  CONSTRAINT customer_nationkey_fk FOREIGN KEY (C_NATIONKEY) REFERENCES NATION(N_NATIONKEY)
) COMMENT 'Customer information including demographics and account details';

CREATE TABLE IF NOT EXISTS PARTSUPP (
  PS_PARTKEY	BIGINT NOT NULL COMMENT 'Foreign key to the part being supplied',
  PS_SUPPKEY	BIGINT NOT NULL COMMENT 'Foreign key to the supplier providing the part',
  PS_AVAILQTY	BIGINT NOT NULL COMMENT 'Available quantity of this part from this supplier',
  PS_SUPPLYCOST	DECIMAL(12,2) NOT NULL COMMENT 'Cost to supply this part from this supplier',
  PS_COMMENT	STRING COMMENT 'General comments about this part-supplier relationship',
  CONSTRAINT partsupp_pk PRIMARY KEY (PS_PARTKEY, PS_SUPPKEY),
  CONSTRAINT partsupp_partkey_fk FOREIGN KEY (PS_PARTKEY) REFERENCES PART(P_PARTKEY),
  CONSTRAINT partsupp_suppkey_fk FOREIGN KEY (PS_SUPPKEY) REFERENCES SUPPLIER(S_SUPPKEY)
) COMMENT 'Association of parts with suppliers including available quantity and cost';

CREATE TABLE IF NOT EXISTS ORDERS (
  O_ORDERKEY	BIGINT NOT NULL COMMENT 'Unique identifier for each order',
  O_CUSTKEY	BIGINT NOT NULL COMMENT 'Foreign key to the customer who placed the order',
  O_ORDERSTATUS	STRING NOT NULL COMMENT 'Status of the order (F=Fulfilled, O=Open, P=Partial)',
  O_TOTALPRICE	DECIMAL(12,2) NOT NULL COMMENT 'Total price of the order including all line items',
  O_ORDERDATE	DATE NOT NULL COMMENT 'Date when the order was placed',
  O_ORDERPRIORITY	STRING NOT NULL COMMENT 'Priority level assigned to the order',
  O_CLERK	STRING NOT NULL COMMENT 'Clerk who processed the order',
  O_SHIPPRIORITY	BIGINT NOT NULL COMMENT 'Shipping priority assigned to the order',
  O_COMMENT	STRING NOT NULL COMMENT 'General comments about the order',
  CONSTRAINT orders_pk PRIMARY KEY (O_ORDERKEY),
  CONSTRAINT orders_custkey_fk FOREIGN KEY (O_CUSTKEY) REFERENCES CUSTOMER(C_CUSTKEY)
) COMMENT 'Customer orders with status, pricing, and processing information';

CREATE TABLE IF NOT EXISTS LINEITEM (
  L_ORDERKEY	BIGINT NOT NULL COMMENT 'Foreign key to the parent order',
  L_PARTKEY	BIGINT NOT NULL COMMENT 'Foreign key to the part being ordered',
  L_SUPPKEY	BIGINT NOT NULL COMMENT 'Foreign key to the supplier of the part',
  L_LINENUMBER	BIGINT NOT NULL COMMENT 'Line number within the order (1, 2, 3, ...)',
  L_QUANTITY	DECIMAL(12,2) NOT NULL COMMENT 'Quantity of the part ordered',
  L_EXTENDEDPRICE	DECIMAL(12,2) NOT NULL COMMENT 'Extended price (quantity * unit price)',
  L_DISCOUNT	DECIMAL(12,2) NOT NULL COMMENT 'Discount percentage applied to this line item',
  L_TAX	DECIMAL(12,2) NOT NULL COMMENT 'Tax percentage applied to this line item',
  L_RETURNFLAG	STRING NOT NULL COMMENT 'Return flag (A=Available, N=Not available, R=Returned)',
  L_LINESTATUS	STRING NOT NULL COMMENT 'Line status (F=Fulfilled, O=Open)',
  L_SHIPDATE	DATE NOT NULL COMMENT 'Date when the line item was shipped',
  L_COMMITDATE	DATE NOT NULL COMMENT 'Date when delivery was committed',
  L_RECEIPTDATE	DATE NOT NULL COMMENT 'Date when the line item was received',
  L_SHIPINSTRUCT	STRING NOT NULL COMMENT 'Shipping instructions for this line item',
  L_SHIPMODE	STRING NOT NULL COMMENT 'Mode of shipping (TRUCK, MAIL, SHIP, AIR, etc.)',
  L_COMMENT	STRING NOT NULL COMMENT 'General comments about this line item',
  CONSTRAINT lineitem_pk PRIMARY KEY (L_ORDERKEY, L_LINENUMBER),
  CONSTRAINT lineitem_orderkey_fk FOREIGN KEY (L_ORDERKEY) REFERENCES ORDERS(O_ORDERKEY),
  CONSTRAINT lineitem_partsupp_fk FOREIGN KEY (L_PARTKEY, L_SUPPKEY) REFERENCES PARTSUPP(PS_PARTKEY, PS_SUPPKEY)
) COMMENT 'Individual line items within orders, including pricing, shipping, and part details';