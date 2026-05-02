CREATE TABLE IF NOT EXISTS projects (
    uuid TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT
);

CREATE TABLE IF NOT EXISTS statuses (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

INSERT OR IGNORE INTO statuses (id, name) VALUES
    (1, 'backlog'),
    (2, 'todo'),
    (3, 'in-progress'),
    (4, 'review'),
    (5, 'done'),
    (6, 'closed');

CREATE TABLE IF NOT EXISTS tasks (
    uuid TEXT PRIMARY KEY,
    project_uuid TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    status_id INTEGER NOT NULL,
    tags TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT,
    FOREIGN KEY (project_uuid) REFERENCES projects(uuid),
    FOREIGN KEY (status_id) REFERENCES statuses(id)
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TEXT NOT NULL,
    updated_at TEXT,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS api_keys (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    label TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS user_project_permissions (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    project_uuid TEXT NOT NULL,
    can_read BOOLEAN NOT NULL DEFAULT TRUE,
    can_write BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (project_uuid) REFERENCES projects(uuid),
    UNIQUE(user_id, project_uuid)
);

CREATE TABLE IF NOT EXISTS _migrations (

    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    applied_at TEXT NOT NULL
);