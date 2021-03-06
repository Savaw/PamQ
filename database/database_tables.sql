CREATE TABLE userinfo (
    username    VARCHAR(50) PRIMARY KEY,
    email       VARCHAR(200) UNIQUE,
    password    VARCHAR(200),
    date_created TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE quiz (
    id              BIGSERIAL PRIMARY KEY,
    creator         VARCHAR(50) NOT NULL REFERENCES userinfo ON DELETE CASCADE,
    name            VARCHAR(200) NOT NULL,
    grading_type    INT NOT NULL,   
    pass_fail       BOOLEAN NOT NULL,
    passing_score   INT,
    not_fail_text   VARCHAR(500),
    fail_text       VARCHAR(500),
    allowed_participations INT NOT NULL,
    date_created    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE quiz_participation (
    id          BIGSERIAL PRIMARY KEY,
    quiz_id     BIGINT NOT NULL REFERENCES quiz ON DELETE CASCADE,
    username    VARCHAR(50) NOT NULL REFERENCES userinfo ON DELETE CASCADE,
    result      VARCHAR(500),
    score       FLOAT,
    pass_fail   BOOLEAN,
    date_created TIMESTAMP WITH TIME ZONE DEFAULT NOW()
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
    answer      VARCHAR(500)
);