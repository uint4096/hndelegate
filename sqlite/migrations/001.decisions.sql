CREATE TABLE IF NOT EXISTS Decisions (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  type TEXT NOT NULL CHECK(type IN ('choice','score','noul')),
  level INTEGER,
  parent INTEGER,
  FOREIGN KEY (parent) REFERENCES Decisions(id)
)
