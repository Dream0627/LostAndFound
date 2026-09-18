USE `laf_db`

DROP TABLE IF EXISTS comments;

DROP TABLE IF EXISTS posts;

DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID',
    username VARCHAR(32) NOT NULL UNIQUE COMMENT '学号或管理员工号',
    name VARCHAR(32) NOT NULL COMMENT '姓名',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希值',
    role ENUM('student', 'postadmin','mainadmin') NOT NULL DEFAULT 'student',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) DEFAULT NULL,
    CONSTRAINT chk_users_username_not_empty CHECK (CHAR_LENGTH(username) BETWEEN 1 AND 32),
    CONSTRAINT chk_users_name_not_empty CHECK (CHAR_LENGTH(name) BETWEEN 1 AND 32)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE posts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '帖子ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '作者ID',
    type ENUM('lost','found') NOT NULL DEFAULT 'lost' COMMENT '帖子类型: lost-丢失寻物, found-寻找失主',
    title VARCHAR(2000) NOT NULL,
    content VARCHAR(2000) NOT NULL,
    image_url VARCHAR(1024) DEFAULT NULL COMMENT '帖子图片的相对路径或URL',
    is_finished BOOLEAN DEFAULT FALSE COMMENT '帖子是否完成',
    status ENUM('pending','approved','rejected') NOT NULL DEFAULT 'pending' COMMENT '审核状态: pending-审核中, approved-已通过, rejected-未通过',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) DEFAULT NULL,
    KEY idx_posts_user_id (user_id),
    KEY idx_posts_created_at (created_at DESC, id DESC),
    KEY idx_posts_type (type),
    KEY idx_posts_status (status), 
    CONSTRAINT fk_posts_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT chk_posts_content_not_empty CHECK (CHAR_LENGTH(content) BETWEEN 1 AND 2000)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE comments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '评论ID',
    post_id BIGINT UNSIGNED NOT NULL COMMENT '所属帖子ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '评论作者ID',
    content VARCHAR(1000) NOT NULL COMMENT '评论内容',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) DEFAULT NULL,
    KEY idx_comments_post_id (post_id),
    KEY idx_comments_user_id (user_id),
    KEY idx_comments_created_at (created_at DESC, id DESC),
    CONSTRAINT fk_comments_post
        FOREIGN KEY (post_id) REFERENCES posts(id)
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT fk_comments_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT chk_comments_content_not_empty CHECK (CHAR_LENGTH(content) BETWEEN 1 AND 1000)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
