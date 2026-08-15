-- name: GetDocumentByID :one
SELECT id, path, content
FROM documents
WHERE id = ?;

-- name: ListDocumentAttributes :many
SELECT name, value
FROM document_attributes
WHERE document_id = ?
ORDER BY name;

-- name: UpsertDocument :exec
INSERT INTO documents (id, path, content)
VALUES (?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    path = excluded.path,
    content = excluded.content;

-- name: DeleteDocumentAttributes :exec
DELETE FROM document_attributes
WHERE document_id = ?;

-- name: InsertDocumentAttribute :exec
INSERT INTO document_attributes (document_id, name, value)
VALUES (?, ?, ?);

-- name: DeleteAllDocuments :exec
DELETE FROM documents;
