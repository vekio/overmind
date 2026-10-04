CREATE TABLE IF NOT EXISTS notes (
 id TEXT PRIMARY KEY,
 type TEXT NOT NULL,
 path TEXT NOT NULL,
 created_at TEXT NOT NULL,
 updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS habits (
 note_id TEXT PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
 title TEXT NOT NULL,
 target_amount REAL NOT NULL CHECK(target_amount > 0),
 unit TEXT NOT NULL,
 period TEXT NOT NULL CHECK(period IN ('day', 'week', 'month'))
);
CREATE TABLE IF NOT EXISTS tags (
 name TEXT PRIMARY KEY
);
CREATE TABLE IF NOT EXISTS note_tags (
 note_id TEXT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
 tag TEXT NOT NULL REFERENCES tags(name),
 position INTEGER NOT NULL CHECK(position >= 0),
 PRIMARY KEY(note_id, tag),
 UNIQUE(note_id, position)
);
CREATE INDEX IF NOT EXISTS note_tags_tag ON note_tags(tag);
CREATE TABLE IF NOT EXISTS persons (
 note_id TEXT PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
 name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS groups (
 name TEXT PRIMARY KEY
);
CREATE TABLE IF NOT EXISTS person_groups (
 person_id TEXT NOT NULL REFERENCES persons(note_id) ON DELETE CASCADE,
 group_name TEXT NOT NULL REFERENCES groups(name),
 position INTEGER NOT NULL CHECK(position >= 0),
 PRIMARY KEY(person_id, group_name),
 UNIQUE(person_id, position)
);
CREATE INDEX IF NOT EXISTS person_groups_group ON person_groups(group_name);
CREATE TABLE IF NOT EXISTS bookmarks (
 note_id TEXT PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
 url TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS inbox (
 note_id TEXT PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
 content TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS pages (
 note_id TEXT PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
 title TEXT NOT NULL,
 area TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS journals (
 note_id TEXT PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
 date TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS notes_updated ON notes(updated_at DESC, id);
CREATE INDEX IF NOT EXISTS notes_type_updated ON notes(type, updated_at DESC, id);
