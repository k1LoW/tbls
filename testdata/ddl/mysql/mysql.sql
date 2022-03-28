DROP TRIGGER IF EXISTS update_posts_updated;
DROP VIEW IF EXISTS post_comments;
DROP TABLE IF EXISTS `hyphen-table`;
DROP TABLE IF EXISTS CamelizeTable;
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS comment_stars;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS user_options;
DROP TABLE IF EXISTS users;
DROP PROCEDURE IF EXISTS GetAllComments;
DROP FUNCTION IF EXISTS CustomerLevel;

CREATE TABLE users (
  id int PRIMARY KEY AUTO_INCREMENT,
  username varchar (50) UNIQUE NOT NULL,
  password varchar (50) NOT NULL,
  email varchar (355) UNIQUE NOT NULL COMMENT 'ex. user@example.com',
  created timestamp NOT NULL,
  updated timestamp
) COMMENT = 'Users table' AUTO_INCREMENT = 100;

CREATE TABLE user_options (
  user_id int PRIMARY KEY,
  show_email boolean NOT NULL DEFAULT false,
  created timestamp NOT NULL,
  updated timestamp,
  UNIQUE(user_id),
  CONSTRAINT user_options_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE NO ACTION ON DELETE CASCADE
) COMMENT = 'User options table';

CREATE TABLE posts (
  id bigint AUTO_INCREMENT,
  user_id int NOT NULL,
  title varchar (255) NOT NULL DEFAULT 'Untitled',
  body text NOT NULL,
  post_type enum('public', 'private', 'draft')  NOT NULL COMMENT 'public/private/draft',
  created datetime NOT NULL,
  updated datetime,
  CONSTRAINT posts_id_pk PRIMARY KEY(id),
  CONSTRAINT posts_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE NO ACTION ON DELETE CASCADE,
  UNIQUE(user_id, title)
) COMMENT = 'Posts table';
CREATE INDEX posts_user_id_idx ON posts(id) USING BTREE;

CREATE TABLE comments (
  id bigint AUTO_INCREMENT,
  post_id bigint NOT NULL,
  user_id int NOT NULL,
  comment text NOT NULL COMMENT 'Comment\nMulti-line\r\ncolumn\rcomment',
  post_id_desc bigint GENERATED ALWAYS AS (post_id * -1) VIRTUAL,
  created datetime NOT NULL,
  updated datetime,
  CONSTRAINT comments_id_pk PRIMARY KEY(id),
  CONSTRAINT comments_post_id_fk FOREIGN KEY(post_id) REFERENCES posts(id),
  CONSTRAINT comments_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id),
  UNIQUE(post_id, user_id)
) COMMENT = 'Comments\nMulti-line\r\ntable\rcomment';
CREATE INDEX comments_post_id_user_id_idx ON comments(post_id, user_id) USING HASH;

CREATE TABLE comment_stars (
  id bigint AUTO_INCREMENT,
  user_id int NOT NULL,
  comment_post_id bigint NOT NULL,
  comment_user_id int NOT NULL,
  created timestamp NOT NULL,
  updated timestamp,
  CONSTRAINT comment_stars_id_pk PRIMARY KEY(id),
  CONSTRAINT comment_stars_user_id_post_id_fk FOREIGN KEY(comment_post_id, comment_user_id) REFERENCES comments(post_id, user_id),
  CONSTRAINT comment_stars_user_id_fk FOREIGN KEY(comment_user_id) REFERENCES users(id),
  UNIQUE(user_id, comment_post_id, comment_user_id)
);

CREATE TABLE logs (
  id bigint PRIMARY KEY AUTO_INCREMENT,
  user_id int NOT NULL,
  post_id bigint,
  comment_id bigint,
  comment_star_id bigint,
  payload text,
  created datetime NOT NULL
) COMMENT = 'Auditログ';

CREATE VIEW post_comments AS (
  SELECT c.id, p.title, u2.username AS post_user, c.comment, u2.username AS comment_user, c.created, c.updated
  FROM posts AS p
  LEFT JOIN comments AS c on p.id = c.post_id
  LEFT JOIN users AS u on u.id = p.user_id
  LEFT JOIN users AS u2 on u2.id = c.user_id
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

CREATE TRIGGER update_posts_updated BEFORE UPDATE ON posts
  FOR EACH ROW
  SET NEW.updated = CURRENT_TIMESTAMP();

DELIMITER //
CREATE PROCEDURE GetAllComments()
BEGIN
	SELECT * FROM comments;
END//
DELIMITER ;

DELIMITER $$
CREATE FUNCTION CustomerLevel(
	credit DECIMAL(10,2)
) 
RETURNS VARCHAR(20)
DETERMINISTIC
BEGIN
    DECLARE customerLevel VARCHAR(20);

    IF credit > 50000 THEN
		SET customerLevel = 'PLATINUM';
    ELSEIF (credit >= 50000 AND 
			credit <= 10000) THEN
        SET customerLevel = 'GOLD';
    ELSEIF credit < 10000 THEN
        SET customerLevel = 'SILVER';
    END IF;
	-- return the customer level
	RETURN (customerLevel);
END$$
DELIMITER ;
