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
  id SERIAL PRIMARY KEY,
  username VARCHAR (50) UNIQUE NOT NULL,
  password VARCHAR (50) NOT NULL,
  email VARCHAR (355) UNIQUE NOT NULL,
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP
);
COMMENT ON TABLE users is 'Users table';
COMMENT ON COLUMN users.email is 'ex. user@example.com';

CREATE TABLE posts (
  id BIGSERIAL NOT NULL,
  user_id INT NOT NULL,
  title VARCHAR (255) NOT NULL,
  body TEXT NOT NULL,
  post_type post_types NOT NULL,
  labels VARCHAR (50)[],
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT posts_id_pk PRIMARY KEY(id),
  CONSTRAINT posts_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) MATCH SIMPLE ON UPDATE NO ACTION ON DELETE CASCADE,
  UNIQUE(user_id, title)
);
COMMENT ON TABLE posts is 'Posts table';
COMMENT ON COLUMN posts.post_type is 'public/private/draft';

CREATE INDEX posts_user_id_idx ON posts USING btree(user_id);

CREATE TABLE comments (
  id BIGSERIAL NOT NULL,
  post_id int NOT NULL,
  user_id int NOT NULL,
  comment TEXT NOT NULL,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT comments_id_pk PRIMARY KEY(id),
  CONSTRAINT comments_post_id_fk FOREIGN KEY(post_id) REFERENCES posts(id) MATCH SIMPLE,
  CONSTRAINT comments_user_id_fk FOREIGN KEY(user_id) REFERENCES users(id) MATCH SIMPLE,
  UNIQUE(post_id, user_id)
);
CREATE INDEX comments_post_id_user_id_idx ON comments USING btree(post_id, user_id);

CREATE TABLE comment_stars (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  user_id int NOT NULL,
  comment_post_id int NOT NULL,
  comment_user_id int NOT NULL,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT comment_stars_user_id_post_id_fk FOREIGN KEY(comment_post_id, comment_user_id) REFERENCES comments(post_id, user_id) MATCH SIMPLE,
  CONSTRAINT comment_stars_user_id_fk FOREIGN KEY(comment_user_id) REFERENCES users(id) MATCH SIMPLE,
  UNIQUE(user_id, comment_post_id, comment_user_id)
);
