DROP TRIGGER IF EXISTS update_posts_updated;
DROP VIEW IF EXISTS post_comment;
DROP TABLE IF EXISTS `hyphen-table`;
DROP TABLE IF EXISTS CamelizeTable;
DROP TABLE IF EXISTS log;
DROP TABLE IF EXISTS comment_star;
DROP TABLE IF EXISTS comment;
DROP TABLE IF EXISTS post;
DROP TABLE IF EXISTS user_option;
DROP TABLE IF EXISTS user;

CREATE TABLE user (
  id int PRIMARY KEY AUTO_INCREMENT,
  username varchar (50) UNIQUE NOT NULL,
  password varchar (50) NOT NULL,
  email varchar (355) UNIQUE NOT NULL COMMENT 'ex. user@example.com',
  created timestamp NOT NULL,
  updated timestamp
) COMMENT = 'User table' AUTO_INCREMENT = 100;

CREATE TABLE user_option (
  user_id int PRIMARY KEY,
  show_email boolean NOT NULL DEFAULT false,
  created timestamp NOT NULL,
  updated timestamp,
  UNIQUE(user_id),
  CONSTRAINT user_option_user_id_fk FOREIGN KEY(user_id) REFERENCES user(id) ON UPDATE NO ACTION ON DELETE CASCADE
) COMMENT = 'User option table';

CREATE TABLE post (
  id bigint AUTO_INCREMENT,
  user_id int NOT NULL,
  title varchar (255) NOT NULL,
  body text NOT NULL,
  post_type enum('public', 'private', 'draft')  NOT NULL COMMENT 'public/private/draft',
  created datetime NOT NULL,
  updated datetime,
  CONSTRAINT post_id_pk PRIMARY KEY(id),
  UNIQUE(user_id, title)
) COMMENT = 'Post table';
CREATE INDEX post_user_id_idx ON post(id) USING BTREE;

CREATE TABLE comment (
  id bigint AUTO_INCREMENT,
  post_id bigint NOT NULL,
  user_id int NOT NULL,
  comment text NOT NULL COMMENT 'Comment\nMulti-line\r\ncolumn\rcomment',
  created datetime NOT NULL,
  updated datetime,
  CONSTRAINT comment_id_pk PRIMARY KEY(id),
  UNIQUE(post_id, user_id)
) COMMENT = 'Comment\nMulti-line\r\ntable\rcomment';
CREATE INDEX comment_post_id_user_id_idx ON comment(post_id, user_id) USING HASH;

CREATE TABLE comment_star (
  id bigint AUTO_INCREMENT,
  user_id int NOT NULL,
  comment_post_id bigint NOT NULL,
  comment_user_id int NOT NULL,
  created timestamp NOT NULL,
  updated timestamp,
  CONSTRAINT comment_star_id_pk PRIMARY KEY(id),
  UNIQUE(user_id, comment_post_id, comment_user_id)
);

CREATE TABLE log (
  id bigint PRIMARY KEY AUTO_INCREMENT,
  user_id int NOT NULL,
  post_id bigint,
  comment_id bigint,
  comment_star_id bigint,
  payload text,
  created datetime NOT NULL
) COMMENT = 'Auditログ';

CREATE VIEW post_comment AS (
  SELECT c.id, p.title, u2.username AS post_user, c.comment, u2.username AS comment_user, c.created, c.updated
  FROM post AS p
  LEFT JOIN comment AS c on p.id = c.post_id
  LEFT JOIN user AS u on u.id = p.user_id
  LEFT JOIN user AS u2 on u2.id = c.user_id
);

CREATE TABLE CamelizeTable (
  id bigint PRIMARY KEY AUTO_INCREMENT,
  created datetime NOT NULL
);

CREATE TABLE `hyphen-table` (
  id bigint PRIMARY KEY AUTO_INCREMENT,
  `hyphen-column` text NOT NULL,
  created datetime NOT NULL
);

CREATE TRIGGER update_posts_updated BEFORE UPDATE ON post
  FOR EACH ROW
  SET NEW.updated = CURRENT_TIMESTAMP();
