DELETE FROM
    presence_log
WHERE
    user_id = $1
    AND location_id = $2
    AND active = $3
