-- +goose Up
CREATE TABLE documents (
    id TEXT PRIMARY KEY NOT NULL,
    path TEXT NOT NULL UNIQUE,
    content BLOB NOT NULL
);

CREATE TABLE document_attributes (
    document_id TEXT NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (document_id, name),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);

CREATE INDEX document_attributes_name_value_idx
    ON document_attributes(name, value);

-- +goose Down
DROP TABLE document_attributes;
DROP TABLE documents;
