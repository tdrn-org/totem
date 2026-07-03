SELECT
    a.mac,
    a.name
FROM
    user_device a
WHERE
    a.user_id = $1
