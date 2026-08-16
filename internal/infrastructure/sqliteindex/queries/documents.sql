-- name: GetDocumentByID :one
SELECT id, path, kind, title, created_at
FROM documents
WHERE id = ?;

-- name: ListDocumentAttributes :many
SELECT name, value
FROM document_attributes
WHERE document_id = ?
ORDER BY name;

-- name: UpsertDocument :exec
INSERT INTO documents (id, path, kind, title, created_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    path = excluded.path,
    kind = excluded.kind,
    title = excluded.title,
    created_at = excluded.created_at;

-- name: DeleteDocumentAttributes :exec
DELETE FROM document_attributes
WHERE document_id = ?;

-- name: InsertDocumentAttribute :exec
INSERT INTO document_attributes (document_id, name, value)
VALUES (?, ?, ?);

-- name: DeleteAllDocuments :exec
DELETE FROM documents;
