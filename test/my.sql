DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;

DROP TABLE IF EXISTS users;

CREATE TABLE users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR (50) UNIQUE NOT NULL,
  password VARCHAR (50) NOT NULL,
  email VARCHAR (355) UNIQUE NOT NULL,
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP
);

CREATE TABLE posts (
  id BIGINT AUTO_INCREMENT,
  user_id INT NOT NULL,
  title VARCHAR (255) NOT NULL,
  body TEXT NOT NULL,
  post_type ENUM('public', 'private', 'draft')  NOT NULL,
  created DATETIME NOT NULL,
  updated DATETIME,
  CONSTRAINT posts_id_pk PRIMARY KEY(id),
  CONSTRAINT posts_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) MATCH SIMPLE ON UPDATE NO ACTION ON DELETE CASCADE
);
CREATE INDEX posts_user_id_idx ON posts(id);

CREATE TABLE comments (
  id bigint AUTO_INCREMENT,
  post_id BIGINT NOT NULL,
  user_id INT NOT NULL,
  comment TEXT NOT NULL,
  created DATETIME NOT NULL,
  updated DATETIME,
  CONSTRAINT comments_id_pk PRIMARY KEY(id),
  CONSTRAINT comments_post_id_fk FOREIGN KEY(post_id) REFERENCES posts(id) MATCH SIMPLE,
  CONSTRAINT comments_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) MATCH SIMPLE
);
CREATE INDEX comments_post_id_user_id_idx ON comments(post_id, user_id) USING BTREE;
