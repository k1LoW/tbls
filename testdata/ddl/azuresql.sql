DROP VIEW IF EXISTS order_summary;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS user_options;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
  id int NOT NULL IDENTITY(1,1) PRIMARY KEY,
  username varchar(50) UNIQUE NOT NULL CHECK(LEN(username) > 4),
  email varchar(355) UNIQUE NOT NULL,
  created date NOT NULL,
  updated date
);

EXEC sys.sp_addextendedproperty  @name=N'MS_Description'
                                ,@value=N'Users table'
                                ,@level0type=N'SCHEMA'
                                ,@level0name=N'dbo'
                                ,@level1type=N'TABLE'
                                ,@level1name=N'users'
                                WITH RESULT SETS NONE;

EXEC sys.sp_addextendedproperty  @name=N'MS_Description'
                                ,@value=N'ex. user@example.com'
                                ,@level0type=N'SCHEMA'
                                ,@level0name=N'dbo'
                                ,@level1type=N'TABLE'
                                ,@level1name=N'users'
                                ,@level2type=N'COLUMN'
                                ,@level2name=N'email'
                                WITH RESULT SETS NONE;

CREATE TABLE user_options (
  user_id int PRIMARY KEY,
  show_email bit NOT NULL DEFAULT 0,
  created date NOT NULL,
  updated date,
  CONSTRAINT user_options_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE NO ACTION ON DELETE CASCADE
);

EXEC sys.sp_addextendedproperty  @name=N'MS_Description'
                                ,@value=N'User options table'
                                ,@level0type=N'SCHEMA'
                                ,@level0name=N'dbo'
                                ,@level1type=N'TABLE'
                                ,@level1name=N'user_options'
                                WITH RESULT SETS NONE;

CREATE TABLE products (
  id int NOT NULL IDENTITY(1,1) PRIMARY KEY,
  name nvarchar(255) NOT NULL,
  price decimal(10,2) NOT NULL,
  stock_quantity int NOT NULL DEFAULT 0,
  created date NOT NULL
);

EXEC sys.sp_addextendedproperty  @name=N'MS_Description'
                                ,@value=N'Product catalog'
                                ,@level0type=N'SCHEMA'
                                ,@level0name=N'dbo'
                                ,@level1type=N'TABLE'
                                ,@level1name=N'products'
                                WITH RESULT SETS NONE;

EXEC sys.sp_addextendedproperty  @name=N'MS_Description'
                                ,@value=N'Unit price in USD'
                                ,@level0type=N'SCHEMA'
                                ,@level0name=N'dbo'
                                ,@level1type=N'TABLE'
                                ,@level1name=N'products'
                                ,@level2type=N'COLUMN'
                                ,@level2name=N'price'
                                WITH RESULT SETS NONE;

CREATE TABLE orders (
  id int NOT NULL IDENTITY(1,1) PRIMARY KEY,
  user_id int NOT NULL,
  total_amount decimal(12,2) NOT NULL,
  status varchar(20) NOT NULL DEFAULT 'pending',
  created date NOT NULL,
  updated date,
  CONSTRAINT orders_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE NO ACTION ON DELETE CASCADE
);

EXEC sys.sp_addextendedproperty  @name=N'MS_Description'
                                ,@value=N'Customer orders'
                                ,@level0type=N'SCHEMA'
                                ,@level0name=N'dbo'
                                ,@level1type=N'TABLE'
                                ,@level1name=N'orders'
                                WITH RESULT SETS NONE;

EXEC sys.sp_addextendedproperty  @name=N'MS_Description'
                                ,@value=N'pending/processing/shipped/delivered'
                                ,@level0type=N'SCHEMA'
                                ,@level0name=N'dbo'
                                ,@level1type=N'TABLE'
                                ,@level1name=N'orders'
                                ,@level2type=N'COLUMN'
                                ,@level2name=N'status'
                                WITH RESULT SETS NONE;

CREATE TABLE order_items (
  id int NOT NULL IDENTITY(1,1) PRIMARY KEY,
  order_id int NOT NULL,
  product_id int NOT NULL,
  quantity int NOT NULL CHECK(quantity > 0),
  unit_price decimal(10,2) NOT NULL,
  created date NOT NULL,
  CONSTRAINT order_items_order_id_fk FOREIGN KEY(order_id) REFERENCES orders(id) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT order_items_product_id_fk FOREIGN KEY(product_id) REFERENCES products(id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  UNIQUE(order_id, product_id)
);

CREATE INDEX order_items_order_id_product_id_idx ON order_items (order_id, product_id);

CREATE VIEW order_summary AS (
  SELECT o.id, u.username, o.total_amount, o.status, o.created
  FROM orders AS o
  LEFT JOIN users AS u ON u.id = o.user_id
);
