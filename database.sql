CREATE TABLE tokens (
    id MEDIUMINT NOT NULL AUTO_INCREMENT, 
    token char(65) NOT NULL,
    type varchar(50) NOT NULL, 
    dbaddr varchar(100) NOT NULL, 
    dbname varchar(30) NOT NULL, 
    username varchar(100) NOT NULL, 
    password blob NOT NULL,
    PRIMARY KEY(id),
    UNIQUE KEY (token),
    INDEX token_ind(token, id)
    ) ENGINE=INNODB;

CREATE TABLE bookmarks (
    token_id MEDIUMINT NOT NULL,
    dbname varchar(100) NOT NULL,
    INDEX token_ind(token_id),
    FOREIGN KEY (token_id)
        REFERENCES tokens(id)
        ON DELETE CASCADE
    ) ENGINE=INNODB;