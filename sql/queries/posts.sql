

-- name: CreatePost :one
INSERT INTO posts(id, updated_at, title, url, description, published_at, feed_id)
VALUES($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPostsForUser :many
SELECT 
    posts.title, 
    posts.description,
    posts.url,
    posts.published_at
FROM posts
JOIN feed_follows ON posts.feed_id = feed_follows.feed_id
JOIN users ON feed_follows.user_id = users.id
WHERE users.name = $1
ORDER BY posts.created_at DESC
LIMIT $2;