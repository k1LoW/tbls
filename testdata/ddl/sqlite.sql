PRAGMA foreign_keys = ON;
DROP TRIGGER IF EXISTS update_posts_updated;
DROP VIEW IF EXISTS post_comments;
DROP TABLE IF EXISTS access_log;
DROP TABLE IF EXISTS syslog;
DROP TABLE IF EXISTS check_constraints;
DROP TABLE IF EXISTS 'hyphen-table';
DROP TABLE IF EXISTS CamelizeTable;
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS comment_stars;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS user_options;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE NOT NULL CHECK(length(username) > 4),
  password TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  created NUMERIC NOT NULL,
  updated NUMERIC
);
CREATE UNIQUE INDEX users_username_key ON users(username);

CREATE TABLE user_options (
  user_id INTEGER PRIMARY KEY,
  show_email INTEGER NOT NULL DEFAULT 0,
  created NUMERIC NOT NULL,
  updated NUMERIC,
  CONSTRAINT user_options_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) MATCH NONE ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE TABLE posts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  created NUMERIC NOT NULL,
  updated NUMERIC,
  CONSTRAINT posts_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) MATCH NONE ON UPDATE NO ACTION ON DELETE CASCADE
);
CREATE INDEX posts_user_id_idx ON posts(user_id);

CREATE TABLE comments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  comment TEXT NOT NULL,
  created NUMERIC NOT NULL,
  updated NUMERIC,
  CONSTRAINT comments_post_id_fk FOREIGN KEY(post_id) REFERENCES posts(id),
  CONSTRAINT comments_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id),
  UNIQUE(post_id, user_id)
);
CREATE INDEX comments_post_id_user_id_idx ON comments(post_id, user_id);

CREATE TABLE comment_stars (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  comment_post_id INTEGER NOT NULL,
  comment_user_id INTEGER NOT NULL,
  created NUMERIC NOT NULL,
  updated NUMERIC,
  CONSTRAINT comment_stars_user_id_post_id_fk FOREIGN KEY(comment_post_id, comment_user_id) REFERENCES comments(post_id, user_id),
  CONSTRAINT comment_stars_user_id_fk FOREIGN KEY(comment_user_id) REFERENCES users(id),
  UNIQUE(user_id, comment_post_id, comment_user_id)
);

CREATE TABLE logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  post_id INTEGER,
  comment_id INTEGER,
  comment_star_id INTEGER,
  payload TEXT,
  created NUMERIC NOT NULL
);

CREATE VIEW post_comments AS
  SELECT c.id, p.title, u2.username AS post_user, c.comment, u2.username AS comment_user, c.created, c.updated
  FROM posts AS p
  LEFT JOIN comments AS c on p.id = c.post_id
  LEFT JOIN users AS u on u.id = p.user_id
  LEFT JOIN users AS u2 on u2.id = c.user_id
;

CREATE TABLE CamelizeTable (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created NUMERIC NOT NULL
);

CREATE TABLE 'hyphen-table' (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  'hyphen-column' TEXT NOT NULL,
  created NUMERIC NOT NULL
);

CREATE TRIGGER update_posts_updated AFTER UPDATE ON posts FOR EACH ROW
BEGIN
  UPDATE posts SET updated = current_timestamp WHERE id = OLD.id;
END;

CREATE TABLE check_constraints (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  col TEXT CHECK(length(col) > 4),
  brackets TEXT UNIQUE NOT NULL CHECK(((length(brackets) > 4))),
  checkcheck TEXT UNIQUE NOT NULL CHECK(length(checkcheck) > 4),
  downcase TEXT UNIQUE NOT NULL check(length(downcase) > 4),
  nl TEXT UNIQUE NOT
    NULL check(length(nl) > 4 OR
      nl != 'ln')
);

CREATE VIRTUAL TABLE syslog USING fts3(logs);
CREATE VIRTUAL TABLE access_log USING fts4(logs);


-- FTS5 tables for full-text search testing
CREATE VIRTUAL TABLE search_posts USING fts5(
  title,
  body,
  content='posts',
  content_rowid='id'
);

CREATE VIRTUAL TABLE search_comments USING fts5(
  comment,
  tokenize='porter unicode61'
);

CREATE VIRTUAL TABLE search_logs USING fts5(
  payload,
  prefix='2 3',
  tokenize='trigram'
);

-- FTS5 with auxiliary columns (not indexed)
CREATE VIRTUAL TABLE search_users USING fts5(
  username,
  email UNINDEXED,
  detail=full
);

-- Insert some test data for FTS5
INSERT INTO users (username, password, email, created) VALUES
  ('alice', 'pass1', 'alice@example.com', datetime('now')),
  ('bobsmith', 'pass2', 'bob@example.com', datetime('now')),
  ('charlie', 'pass3', 'charlie@example.com', datetime('now'));

INSERT INTO posts (user_id, title, body, created) VALUES
  (1, 'Hello World', 'This is my first post about SQLite FTS5', datetime('now')),
  (2, 'Database Search', 'Full-text search is amazing with FTS5', datetime('now')),
  (3, 'Performance Tips', 'Optimize your queries using indexes', datetime('now'));

INSERT INTO search_comments (comment) VALUES
  ('Great article about full-text search!'),
  ('FTS5 is much faster than FTS3'),
  ('Thanks for sharing this information');

INSERT INTO search_logs (payload) VALUES
  ('User login successful'),
  ('Database connection established'),
  ('Search query executed');

INSERT INTO search_users (username, email) VALUES
  ('alice', 'alice@example.com'),
  ('bobsmith', 'bob@example.com'),
  ('charlie', 'charlie@example.com');

-- Triggers to keep FTS5 in sync with content table
CREATE TRIGGER posts_fts_insert AFTER INSERT ON posts BEGIN
  INSERT INTO search_posts(rowid, title, body) VALUES (new.id, new.title, new.body);
END;

CREATE TRIGGER posts_fts_delete AFTER DELETE ON posts BEGIN
  INSERT INTO search_posts(search_posts, rowid, title, body) VALUES('delete', old.id, old.title, old.body);
END;

CREATE TRIGGER posts_fts_update AFTER UPDATE ON posts BEGIN
  INSERT INTO search_posts(search_posts, rowid, title, body) VALUES('delete', old.id, old.title, old.body);
  INSERT INTO search_posts(rowid, title, body) VALUES (new.id, new.title, new.body);
END;
