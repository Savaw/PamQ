CREATE TABLE userinfo (
    username    VARCHAR(50) PRIMARY KEY,
    email       VARCHAR(200) UNIQUE,
    password    VARCHAR(200)
);

CREATE TABLE quiz (
    id          BIGSERIAL PRIMARY KEY,
    creator     VARCHAR(50) NOT NULL REFERENCES userinfo ON DELETE CASCADE,
    name        VARCHAR(200) NOT NULL
);

CREATE TABLE quiz_participation (
    id          BIGSERIAL PRIMARY KEY,
    quiz_id     BIGINT NOT NULL REFERENCES quiz ON DELETE CASCADE,
    username    VARCHAR(50) NOT NULL REFERENCES userinfo ON DELETE CASCADE,
    result      VARCHAR(500)
);

CREATE TABLE question (
    id          BIGSERIAL PRIMARY KEY,
    quiz_id     BIGINT NOT NULL REFERENCES quiz ON DELETE CASCADE,
    type        INT NOT NULL,
    statement   VARCHAR(500) NOT NULL,
    option1     VARCHAR(500),
    option2     VARCHAR(500),
    option3     VARCHAR(500),
    option4     VARCHAR(500),
    answer      INT
);