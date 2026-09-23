-- +goose Up
CREATE TABLE notes (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('page', 'journal', 'inbox', 'bookmark')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE note_attributes (
    note_id TEXT NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (note_id, name),
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
);

CREATE TABLE note_tags (
    note_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    position INTEGER NOT NULL,
    PRIMARY KEY (note_id, tag),
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
);

CREATE INDEX notes_kind_id_idx ON notes(kind, id);
CREATE INDEX note_attributes_name_value_note_id_idx
    ON note_attributes(name, value, note_id);
CREATE INDEX note_tags_tag_note_id_idx ON note_tags(tag, note_id);

-- +goose Down
DROP INDEX note_tags_tag_note_id_idx;
DROP INDEX note_attributes_name_value_note_id_idx;
DROP INDEX notes_kind_id_idx;
DROP TABLE note_tags;
DROP TABLE note_attributes;
DROP TABLE notes;
