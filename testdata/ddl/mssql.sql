DROP TRIGGER IF EXISTS update_users_updated;
DROP TRIGGER IF EXISTS update_posts_updated;
DROP TABLE IF EXISTS administrator.blogs;
DROP VIEW IF EXISTS post_comments;
DROP TABLE IF EXISTS "hyphen-table";
DROP TABLE IF EXISTS "CamelizeTable";
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS comment_stars;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;
DROP TYPE IF EXISTS post_types;
DROP TABLE IF EXISTS user_options;
DROP TABLE IF EXISTS users;
DROP SCHEMA IF EXISTS administrator;
DROP FUNCTION IF EXISTS get_user;
DROP PROC IF EXISTS What_DB_is_that;

CREATE TABLE users (
  id int NOT NULL IDENTITY(1,1) PRIMARY KEY,
  username varchar (50) UNIQUE NOT NULL CHECK(LEN(username) > 4),
  password varchar (50) NOT NULL,
  email varchar (355) UNIQUE NOT NULL,
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

EXEC sys.sp_addextendedproperty  @name=N'MS_Description'
                                ,@value=N'long long long long long long long long long long long long long long long long long long long long long description'
                                ,@level0type=N'SCHEMA'
                                ,@level0name=N'dbo'
                                ,@level1type=N'TABLE'
                                ,@level1name=N'users'
                                ,@level2type=N'COLUMN'
                                ,@level2name=N'password'
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

CREATE TABLE posts (
  id int NOT NULL,
  user_id int NOT NULL,
  title varchar (255) NOT NULL,
  body text NOT NULL,
  created date NOT NULL,
  updated date,
  CONSTRAINT posts_id_pk PRIMARY KEY(id),
  CONSTRAINT posts_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE NO ACTION ON DELETE CASCADE,
  UNIQUE(user_id, title)
);

CREATE INDEX posts_user_id_idx ON posts (user_id);

CREATE TABLE comments (
  id int NOT NULL,
  post_id int NOT NULL,
  user_id int NOT NULL,
  comment text NOT NULL,
  created date NOT NULL,
  updated date,
  CONSTRAINT comments_id_pk PRIMARY KEY(id),
  CONSTRAINT comments_post_id_fk FOREIGN KEY(post_id) REFERENCES posts(id),
  CONSTRAINT comments_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id),
  UNIQUE(post_id, user_id)
);

CREATE INDEX comments_post_id_user_id_idx ON comments (post_id, user_id);

CREATE TABLE comment_stars (
  id int NOT NULL,
  user_id int NOT NULL,
  comment_post_id int NOT NULL,
  comment_user_id int NOT NULL,
  created date NOT NULL,
  updated date,
  CONSTRAINT comment_stars_user_id_post_id_fk FOREIGN KEY(comment_post_id, comment_user_id) REFERENCES comments(post_id, user_id),
  CONSTRAINT comment_stars_user_id_fk FOREIGN KEY(comment_user_id) REFERENCES users(id),
  UNIQUE(user_id, comment_post_id, comment_user_id)
);

CREATE TABLE logs (
  id int NOT NULL,
  user_id int NOT NULL,
  post_id int,
  comment_id int,
  comment_star_id int,
  payload text,
  created date NOT NULL
);

CREATE VIEW post_comments AS (
  SELECT c.id, p.title, u2.username AS post_user, c.comment, u2.username AS comment_user, c.created, c.updated
  FROM posts AS p
  LEFT JOIN comments AS c on p.id = c.post_id
  LEFT JOIN users AS u on u.id = p.user_id
  LEFT JOIN users AS u2 on u2.id = c.user_id
);

CREATE TABLE "CamelizeTable" (
  id int NOT NULL,
  created date NOT NULL
);

CREATE TABLE "hyphen-table" (
  id int NOT NULL,
  "hyphen-column" text NOT NULL,
  created date NOT NULL
);

CREATE SCHEMA administrator;

CREATE TABLE administrator.blogs (
  id int NOT NULL IDENTITY(1,1) PRIMARY KEY,
  user_id int NOT NULL,
  name text NOT NULL,
  description text,
  created date NOT NULL,
  updated date,
  CONSTRAINT blogs_user_id_fk FOREIGN KEY(user_id) REFERENCES dbo.users(id) ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TRIGGER update_users_updated
ON users
AFTER UPDATE
AS
BEGIN
  UPDATE users SET updated = GETDATE()
  WHERE id = ( SELECT id FROM deleted)
END;

CREATE TRIGGER update_posts_updated
ON posts
AFTER UPDATE
AS
BEGIN
  UPDATE users SET updated = GETDATE()
  WHERE id = ( SELECT user_id FROM deleted)
END;

CREATE FUNCTION get_user (@userid int)
RETURNS TABLE
AS
RETURN
(
  SELECT u.username, u.email
  FROM users AS u
  WHERE u.id = @userid
);

CREATE PROC What_DB_is_that @ID INT
AS
SELECT DB_NAME(@ID) AS ThatDB;
