UPDATE
    user
SET
    name = $1,
    default_location_id = $2
WHERE
    id = $3
