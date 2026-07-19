SELECT
    a.id,
    a.agent,
    a.message,
    a.scheduled_at,
    a.created_at
FROM
    pulse a
ORDER BY
    a.scheduled_at ASC
