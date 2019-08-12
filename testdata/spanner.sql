CREATE TABLE users (
  user_id INT64 NOT NULL,
  username STRING(50) NOT NULL,
  password STRING(50) NOT NULL,
  email STRING(255) NOT NULL,
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (user_id);
CREATE UNIQUE INDEX users_username_idx ON users(username);
CREATE UNIQUE INDEX users_email_idx ON users(email);

CREATE TABLE user_options (
  user_id INT64 NOT NULL,
  show_email BOOL NOT NULL,
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (user_id),
INTERLEAVE IN PARENT users ON DELETE CASCADE;

CREATE TABLE posts (
  user_id INT64 NOT NULL,
  post_id INT64 NOT NULL,
  title STRING(255) NOT NULL,
  body STRING(MAX) NOT NULL,
  image BYTES(MAX) NOT NULL,
  labels ARRAY<STRING(50)>,
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (user_id, post_id),
INTERLEAVE IN PARENT users ON DELETE CASCADE;
CREATE UNIQUE NULL_FILTERED INDEX posts_user_id_title_idx ON posts(user_id, title);
CREATE INDEX posts_user_id_idx ON posts(user_id) STORING (title);

CREATE TABLE comments (
  user_id int64 NOT NULL,
  post_id INT64 NOT NULL,
  comment_id INT64 NOT NULL,
  comment STRING(MAX) NOT NULL,
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY(user_id, post_id, comment_id),
INTERLEAVE IN PARENT posts ON DELETE CASCADE;

CREATE UNIQUE INDEX comments_post_id_user_id_idx ON comments(post_id, user_id);

CREATE TABLE comment_stars (
  user_id INT64 NOT NULL,
  comment_star_id INT64 NOT NULL,
  comment_post_id INT64 NOT NULL,
  comment_user_id INT64 NOT NULL,
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY(user_id, comment_star_id),
INTERLEAVE IN PARENT users ON DELETE CASCADE;
CREATE UNIQUE INDEX comment_stars_idx ON comment_stars(user_id, comment_post_id, comment_user_id);

CREATE TABLE logs (
  log_id INT64 NOT NULL,
  user_id INT64 NOT NULL,
  post_id INT64,
  comment_id INT64,
  comment_star_id INT64,
  payload STRING(MAX),
  created timestamp NOT NULL
) PRIMARY KEY(log_id);

CREATE TABLE CamelizeTable (
  CamelizeTableId INT64 NOT NULL,
  created TIMESTAMP NOT NULL
) PRIMARY KEY(CamelizeTableId);
