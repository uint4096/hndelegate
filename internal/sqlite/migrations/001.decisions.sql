CREATE TABLE IF NOT EXISTS Decisions (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK(type IN ('choice','score','noul')),
    name TEXT NOT NULL,
    instruction TEXT NOT NULL,
    parent INTEGER,
    version INTEGER NOT NULL,
    min_confidence FLOAT,
    UNIQUE (name, version),
    CHECK(
        min_confidence IS NULL
        OR (type = 'choice' AND min_confidence BETWEEN 0 AND 1)
    ),
    FOREIGN KEY (parent) REFERENCES Decisions(id)
);

CREATE TABLE IF NOT EXISTS Choices (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    value TEXT NOT NULL,
    decision INTEGER NOT NULL,
    description TEXT,
    acceptable_probability FLOAT,
    FOREIGN KEY (decision) REFERENCES Decisions(id)
);
