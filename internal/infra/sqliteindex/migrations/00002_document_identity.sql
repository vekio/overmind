-- +goose Up
DROP INDEX documents_kind_path_id_idx;
DROP INDEX document_tags_tag_document_id_idx;
DROP INDEX document_attributes_name_value_idx;
DROP TABLE document_tags;
DROP TABLE document_attributes;
DROP TABLE documents;

CREATE TABLE documents (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    title TEXT NOT NULL,
    area TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE document_attributes (
    document_id TEXT NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (document_id, name),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);

CREATE TABLE document_tags (
    document_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    position INTEGER NOT NULL,
    PRIMARY KEY (document_id, tag),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);

CREATE INDEX document_tags_tag_document_id_idx
    ON document_tags(tag, document_id);

CREATE INDEX document_attributes_name_value_idx
    ON document_attributes(name, value);

CREATE INDEX documents_title_area_kind_id_idx
    ON documents(title COLLATE NOCASE, area, kind, id);

-- +goose Down
DROP INDEX documents_title_area_kind_id_idx;
DROP INDEX document_tags_tag_document_id_idx;
DROP INDEX document_attributes_name_value_idx;
DROP TABLE document_tags;
DROP TABLE document_attributes;
DROP TABLE documents;

CREATE TABLE documents (
    id TEXT PRIMARY KEY NOT NULL,
    path TEXT NOT NULL UNIQUE,
    kind TEXT NOT NULL,
    title TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE document_attributes (
    document_id TEXT NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (document_id, name),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);

CREATE TABLE document_tags (
    document_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    position INTEGER NOT NULL,
    PRIMARY KEY (document_id, tag),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);

CREATE INDEX document_tags_tag_document_id_idx
    ON document_tags(tag, document_id);

CREATE INDEX document_attributes_name_value_idx
    ON document_attributes(name, value);

CREATE INDEX documents_kind_path_id_idx
    ON documents(kind, path, id);
