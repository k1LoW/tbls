DROP VIEW IF EXISTS post_comments;
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS comment_stars;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;
DROP TYPE IF EXISTS post_types;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";
CREATE EXTENSION "uuid-ossp";

CREATE TYPE post_types AS ENUM (
  'public', 'private', 'draft'
);

CREATE TABLE users (
  id serial PRIMARY KEY,
  username varchar (50) UNIQUE NOT NULL,
  password varchar (50) NOT NULL,
  email varchar (355) UNIQUE NOT NULL,
  created timestamp NOT NULL,
  updated timestamp
);
COMMENT ON TABLE users is 'Users table';
COMMENT ON COLUMN users.email is 'ex. user@example.com';

CREATE TABLE posts (
  id bigserial NOT NULL,
  user_id int NOT NULL,
  title varchar (255) NOT NULL,
  body text NOT NULL,
  post_type post_types NOT NULL,
  labels varchar (50)[],
  created timestamp without time zone NOT NULL,
  updated timestamp without time zone,
  CONSTRAINT posts_id_pk PRIMARY KEY(id),
  CONSTRAINT posts_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) MATCH SIMPLE ON UPDATE NO ACTION ON DELETE CASCADE,
  UNIQUE(user_id, title)
);
COMMENT ON TABLE posts is 'Posts table';
COMMENT ON COLUMN posts.post_type is 'public/private/draft';

CREATE INDEX posts_user_id_idx ON posts USING btree(user_id);

CREATE TABLE comments (
  id bigserial NOT NULL,
  post_id bigint NOT NULL,
  user_id int NOT NULL,
  comment text NOT NULL,
  created timestamp without time zone NOT NULL,
  updated timestamp without time zone,
  CONSTRAINT comments_id_pk PRIMARY KEY(id),
  CONSTRAINT comments_post_id_fk FOREIGN KEY(post_id) REFERENCES posts(id) MATCH SIMPLE,
  CONSTRAINT comments_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) MATCH SIMPLE,
  UNIQUE(post_id, user_id)
);
CREATE INDEX comments_post_id_user_id_idx ON comments USING btree(post_id, user_id);

CREATE TABLE comment_stars (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  user_id int NOT NULL,
  comment_post_id bigint NOT NULL,
  comment_user_id int NOT NULL,
  created timestamp without time zone NOT NULL,
  updated timestamp without time zone,
  CONSTRAINT comment_stars_user_id_post_id_fk FOREIGN KEY(comment_post_id, comment_user_id) REFERENCES comments(post_id, user_id) MATCH SIMPLE,
  CONSTRAINT comment_stars_user_id_fk FOREIGN KEY(comment_user_id) REFERENCES users(id) MATCH SIMPLE,
  UNIQUE(user_id, comment_post_id, comment_user_id)
);

CREATE TABLE logs (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  user_id int NOT NULL,
  post_id bigint,
  comment_id bigint,
  comment_star_id uuid,
  payload text,
  created timestamp NOT NULL
);

CREATE VIEW post_comments AS (
  SELECT c.id, p.title, u2.username AS post_user, c.comment, u2.username AS comment_user, c.created, c.updated
  FROM posts AS p
  LEFT JOIN comments AS c on p.id = c.post_id
  LEFT JOIN users AS u on u.id = p.user_id
  LEFT JOIN users AS u2 on u2.id = c.user_id
);
