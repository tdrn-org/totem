INSERT INTO
    presence_log(
        user_id,
        location_id,
        active,
        ip,
        name,
        timestamp
    )
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6    
)
