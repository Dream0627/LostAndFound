USE `laf_db`

DROP TABLE IF EXISTS comments;

DROP TABLE IF EXISTS posts;

DROP TABLE IF EXISTS appeals;
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
    location_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '校园预设地点ID(冗余快照)',
    location_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '地点名称快照(冗余,减少前端查询)',
    supplement VARCHAR(200) NOT NULL DEFAULT '' COMMENT '地点补充说明',
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

-- 申诉表：用户账号被注销(软删除)后无法登录，故由被注销者以 username 公开提交申诉，
-- 超级管理员审核(pending -> approved/rejected)；审核通过后自动恢复对应账号。
DROP TABLE IF EXISTS appeals;

CREATE TABLE appeals (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '申诉ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '申诉人ID',
    reason ENUM('self_regret','wrongful_ban','other') NOT NULL DEFAULT 'other' COMMENT '申诉原因: self_regret-自行注销反悔, wrongful_ban-被误封号请求撤回, other-其他',
    content VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '申诉说明(other 原因必填)',
    status ENUM('pending','approved','rejected') NOT NULL DEFAULT 'pending' COMMENT '审核状态',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_appeals_user_id (user_id),
    KEY idx_appeals_status (status),
    KEY idx_appeals_created_at (created_at DESC, id DESC),
    CONSTRAINT fk_appeals_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户注销申诉表';

-- 对话表：对某帖「申领(found)」/「召领(lost)」后开启，连接发起方与帖子作者，作为双方私聊的容器。
DROP TABLE IF EXISTS finish_requests;

DROP TABLE IF EXISTS messages;

DROP TABLE IF EXISTS conversations;

CREATE TABLE conversations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '对话ID',
    post_id BIGINT UNSIGNED NOT NULL COMMENT '所属帖子ID',
    initiator_id BIGINT UNSIGNED NOT NULL COMMENT '发起方(申领/召领人)ID',
    owner_id BIGINT UNSIGNED NOT NULL COMMENT '帖子作者ID',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_conversations_post_initiator (post_id, initiator_id),
    KEY idx_conversations_initiator_id (initiator_id),
    KEY idx_conversations_owner_id (owner_id),
    KEY idx_conversations_created_at (created_at DESC, id DESC),
    CONSTRAINT fk_conversations_post
        FOREIGN KEY (post_id) REFERENCES posts(id)
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT fk_conversations_initiator
        FOREIGN KEY (initiator_id) REFERENCES users(id)
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT fk_conversations_owner
        FOREIGN KEY (owner_id) REFERENCES users(id)
        ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='申领/召领对话表';

CREATE TABLE messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',
    conversation_id BIGINT UNSIGNED NOT NULL COMMENT '所属对话ID',
    sender_id BIGINT UNSIGNED NOT NULL COMMENT '发送者ID',
    content VARCHAR(1000) NOT NULL COMMENT '消息内容',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_messages_conversation_id (conversation_id),
    KEY idx_messages_sender_id (sender_id),
    KEY idx_messages_created_at (created_at DESC, id DESC),
    CONSTRAINT fk_messages_conversation
        FOREIGN KEY (conversation_id) REFERENCES conversations(id)
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT fk_messages_sender
        FOREIGN KEY (sender_id) REFERENCES users(id)
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT chk_messages_content_not_empty CHECK (CHAR_LENGTH(content) BETWEEN 1 AND 1000)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='对话消息表';

-- 完成寻找申请表：对话任一方发起，另一方处理；同意(agreed)后对应帖子置 is_finished=true。
CREATE TABLE finish_requests (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '完成申请ID',
    conversation_id BIGINT UNSIGNED NOT NULL COMMENT '所属对话ID',
    requester_id BIGINT UNSIGNED NOT NULL COMMENT '发起人ID',
    status ENUM('pending','agreed','rejected') NOT NULL DEFAULT 'pending' COMMENT '申请状态: pending-待处理, agreed-已同意, rejected-已拒绝',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    KEY idx_finish_requests_conversation_id (conversation_id),
    KEY idx_finish_requests_requester_id (requester_id),
    KEY idx_finish_requests_status (status),
    CONSTRAINT fk_finish_requests_conversation
        FOREIGN KEY (conversation_id) REFERENCES conversations(id)
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT fk_finish_requests_requester
        FOREIGN KEY (requester_id) REFERENCES users(id)
        ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='完成寻找申请表';
