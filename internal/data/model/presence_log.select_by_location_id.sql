SELECT
    a.user_id,
    a.active,
    a.ip,
    a.name,
    a.timestamp
FROM
    presence_log a
WHERE
    location_id = $1
ORDER BY
    a.timestamp DESC
