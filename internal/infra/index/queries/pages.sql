-- name: UpsertPage :exec
INSERT INTO pages(note_id,title,area) VALUES(?,?,?)
ON CONFLICT(note_id) DO UPDATE SET title=excluded.title,area=excluded.area;
