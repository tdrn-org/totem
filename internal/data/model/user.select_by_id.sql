SELECT
    a.id,
    a.name,
    a.default_location_id
FROM
    user a
WHERE
    a.user_id = $1
