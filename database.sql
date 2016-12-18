
-- Make sure your mysql process is running first!

DROP DATABASE IF EXISTS `VisAwesome`;
CREATE DATABASE IF NOT EXISTS `VisAwesome`  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `VisAwesome`;

-- Enable client program to communicate with the server using utf8 character set
SET NAMES 'utf8';

DROP TABLE IF EXISTS `Category`;
create table IF NOT EXISTS `Category`(
    `id`            INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`          varchar(48) not null,
    `url`     varchar(256) not null,
    `topic`   varchar(128) not null,
    `status`  enum('NoStatus','Start','Fetching','Done','Error') not null default 'NoStatus',
    UNIQUE(`name`),
    UNIQUE(`url`),
    PRIMARY KEY (`id`)
)DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EXPLAIN `Category`;


DROP TABLE IF EXISTS `Repos`;
create table IF NOT EXISTS `Repos`(
    `id`            INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`          varchar(48) not null,
    `url`           varchar(256) not null,
    `category_id`      INT(10) UNSIGNED,
    `status`  enum('NoStatus','Start','Fetching','Done','Error') not null default 'NoStatus',
    `error_message` varchar(256) not null default '',
    `sub_topic`  varchar(128) not null default '',
    `fork_number`   INT not null default 0,
    `star_number`   INT not null default 0,
    `watch_number`  INT NOT NULL DEFAULT 0,
    `contributors`  INT not null default 0,
    UNIQUE(`name`),
    UNIQUE(`url`),
    PRIMARY KEY (`id`),
    FOREIGN KEY (`category_id`) REFERENCES Category(`id`)
)DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EXPLAIN `Category`;


