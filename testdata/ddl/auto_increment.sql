DROP TABLE IF EXISTS users;

CREATE TABLE users (
  id int PRIMARY KEY AUTO_INCREMENT,
  username varchar (50) UNIQUE NOT NULL,
  password varchar (50) NOT NULL,
  email varchar (355) UNIQUE NOT NULL COMMENT 'ex. user@example.com',
  created timestamp NOT NULL,
  updated timestamp
) COMMENT = 'Users table';
