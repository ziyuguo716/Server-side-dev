CREATE DATABASE mydb;
USE mydb;
CREATE TABLE user (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(191) NOT NULL,
    passhash BINARY(60) NOT NULL,
    user_name VARCHAR(191) NOT NULL,
    first_name VARCHAR(128),
    last_name VARCHAR(128),
    photo_url VARCHAR(191) NOT NULL
);

CREATE index index_username ON user (user_name);
CREATE index index_email ON user (email);
