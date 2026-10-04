-- name: FindNotes :many
SELECT n.id, n.type, n.updated_at,
 CAST(CASE n.type
 WHEN 'habit' THEN h.title
 WHEN 'person' THEN p.name
 WHEN 'bookmark' THEN b.url
 WHEN 'inbox' THEN substr(i.content, 1, 120)
 WHEN 'page' THEN pg.title
 WHEN 'journal' THEN j.date
 ELSE '' END AS TEXT) AS label,
 CAST((SELECT json_group_array(tag) FROM
 (SELECT tag FROM note_tags WHERE note_id = n.id ORDER BY position)) AS TEXT) AS tags
FROM notes n
LEFT JOIN habits h ON h.note_id = n.id
LEFT JOIN persons p ON p.note_id = n.id
LEFT JOIN bookmarks b ON b.note_id = n.id
LEFT JOIN inbox i ON i.note_id = n.id
LEFT JOIN pages pg ON pg.note_id = n.id
LEFT JOIN journals j ON j.note_id = n.id
WHERE (sqlc.arg(note_type) = '' OR n.type = sqlc.arg(note_type))
 AND (sqlc.arg(tag) = '' OR EXISTS
 (SELECT 1 FROM note_tags nt WHERE nt.note_id = n.id AND nt.tag = sqlc.arg(tag)))
ORDER BY n.updated_at DESC, n.id ASC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: ClearNotes :exec
DELETE FROM notes;

-- name: ClearTags :exec
DELETE FROM tags;

-- name: ClearGroups :exec
DELETE FROM groups;

-- name: ClearAllNoteTags :exec
DELETE FROM note_tags;

-- name: ClearAllPersonGroups :exec
DELETE FROM person_groups;

-- name: ClearHabits :exec
DELETE FROM habits;

-- name: ClearPersons :exec
DELETE FROM persons;

-- name: ClearBookmarks :exec
DELETE FROM bookmarks;

-- name: ClearInbox :exec
DELETE FROM inbox;

-- name: ClearPages :exec
DELETE FROM pages;

-- name: ClearJournals :exec
DELETE FROM journals;
