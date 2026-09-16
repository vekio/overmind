-- name: GetDocumentByID :one
SELECT id, path, kind, title, created_at, updated_at
FROM documents
WHERE id = ?;

-- name: ListDocuments :many
SELECT id, path, kind
FROM documents AS document
WHERE (CAST(sqlc.arg(kind) AS TEXT) = '' OR document.kind = CAST(sqlc.arg(kind) AS TEXT))
  AND (CAST(sqlc.arg(path_prefix) AS TEXT) = ''
       OR instr(document.path, CAST(sqlc.arg(path_prefix) AS TEXT)) = 1)
  AND NOT EXISTS (
      SELECT 1
      FROM json_each(CAST(sqlc.arg(tags_json) AS TEXT)) AS requested_tag
      WHERE NOT EXISTS (
          SELECT 1
          FROM document_tags AS document_tag
          WHERE document_tag.document_id = document.id
            AND document_tag.tag = requested_tag.value
      )
  )
ORDER BY document.path, document.id;

-- name: ListDocumentAttributes :many
SELECT name, value
FROM document_attributes
WHERE document_id = ?
ORDER BY name;

-- name: ListDocumentTags :many
SELECT tag
FROM document_tags
WHERE document_id = ?
ORDER BY position;

-- name: UpsertDocument :exec
INSERT INTO documents (id, path, kind, title, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    path = excluded.path,
    kind = excluded.kind,
    title = excluded.title,
    created_at = excluded.created_at,
    updated_at = excluded.updated_at;

-- name: DeleteDocumentAttributes :exec
DELETE FROM document_attributes
WHERE document_id = ?;

-- name: DeleteDocumentTags :exec
DELETE FROM document_tags
WHERE document_id = ?;

-- name: InsertDocumentAttribute :exec
INSERT INTO document_attributes (document_id, name, value)
VALUES (?, ?, ?);

-- name: InsertDocumentTag :exec
INSERT INTO document_tags (document_id, tag, position)
VALUES (?, ?, ?);

-- name: DeleteAllDocuments :exec
DELETE FROM documents;
