CREATE TABLE IF NOT EXISTS agents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    zona TEXT NOT NULL,
    agente TEXT NOT NULL,
    note TEXT NOT NULL,
    tel TEXT NOT NULL,
    email TEXT NOT NULL,
    email_personal TEXT NOT NULL
);
