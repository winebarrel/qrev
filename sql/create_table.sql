CREATE TABLE qrev_history (
    filename       VARCHAR(255) PRIMARY KEY,
    hash           VARCHAR(255) NOT NULL,
    executed_at    VARCHAR(255) NOT NULL,
    execution_time INTEGER      NOT NULL,
    status         VARCHAR(255) NOT NULL,
    last_error     TEXT         NOT NULL
);
