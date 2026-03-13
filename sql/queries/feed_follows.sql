
-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows(id, updated_at, user_id, feed_id)
    VALUES( $1, $2, $3, $4)
    RETURNING * 
)
SELECT 
    inserted_feed_follow.*,
    users.name as user_name,
    feeds.name as feed_name,
    feeds.url as feed_url
FROM inserted_feed_follow
JOIN users ON inserted_feed_follow.user_id = users.id
JOIN feeds ON inserted_feed_follow.feed_id = feeds.id;



-- name: GetFeedFollowsForUser :many
SELECT 
    users.name as user_name,
    feeds.name as feed_name
FROM feed_follows
JOIN users ON users.id = user_id
JOIN feeds ON feeds.id = feed_id
WHERE users.name = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;