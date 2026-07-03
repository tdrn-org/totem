SELECT
    a.location_id,
    a.active,
    a.ip,
    a.name,
    a.timestamp
FROM
    presence_log a
WHERE
    user_id = $1
ORDER BY
    a.timestamp DESC
