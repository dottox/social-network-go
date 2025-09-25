CREATE TABLE IF NOT EXISTS roles (
    id INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    level INT NOT NULL DEFAULT 0,
    description TEXT
);

INSERT INTO roles (id, name, level, description) VALUES
(3, 'admin', 3, 'Administrator with full access'),
(2, 'moderator', 2, 'A moderator can update other users posts'),
(1, 'user', 1, 'A user can create posts and comments');