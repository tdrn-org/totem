SELECT
    a.id,
    a.message,
    a.scheduled_at,
    a.created_at
FROM
    pulse a
WHERE
    agent = $1
ORDER BY
    a.scheduled_at ASC
