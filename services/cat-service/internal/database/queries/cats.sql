-- name: CreateCat :one
INSERT INTO cats (
  owner_id,
  name,
  breed,
  date_of_birth,
  weight,
  color,
  gender,
  photo_url,
  medical_notes,
  dietary_requirements
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetCatByID :one
SELECT * FROM cats
WHERE id = $1
AND deleted_at IS NULL;

-- name: CatByOwnerExists :one
SELECT EXISTS (
  SELECT 1
  FROM cats
  WHERE id = $1
  AND owner_id = $2
  AND deleted_at IS NULL
) AS exists;

-- name: ListCatsByOwner :many
SELECT * FROM cats
WHERE owner_id = $1
AND deleted_at IS NULL;

-- name: UpdateCat :one
UPDATE cats SET
  name = COALESCE($1, name),
  breed = COALESCE($2, breed),
  date_of_birth = COALESCE($3, date_of_birth),
  weight = COALESCE($4, weight),
  color = COALESCE($5, color),
  gender = COALESCE($6, gender),
  photo_url = COALESCE($7, photo_url),
  medical_notes = COALESCE($8, medical_notes),
  dietary_requirements = COALESCE($9, dietary_requirements),
  date_of_death = COALESCE($10, date_of_death),
  updated_at = NOW()
WHERE id = $11
AND owner_id = $12
AND deleted_at IS NULL
RETURNING *;

-- name: ClearDateOfDeath :exec
UPDATE cats
SET date_of_death = NULL
WHERE id = $1
AND owner_id = $2
AND deleted_at IS NULL;

-- name: GetCatsName :one
SELECT name
FROM cats
WHERE id = $1
AND owner_id = $2
AND deleted_at IS NULL;

-- name: GetCatsBirthday :one
SELECT date_of_birth
FROM cats
WHERE id = $1
AND owner_id = $2
AND deleted_at IS NULL;

-- name: SoftDeleteCat :exec
UPDATE cats
SET deleted_at = NOW()
WHERE id = $1
AND owner_id = $2
AND deleted_at IS NULL;

-- name: SoftDeleteCatsByOwner :exec
UPDATE cats
SET deleted_at = NOW()
WHERE owner_id = $1;

-- name: CountCatsByOwner :one
SELECT COUNT(*)
FROM cats
WHERE owner_id = $1
AND deleted_at IS NULL;

-- name: GetCatOwner :one
SELECT owner_id
FROM cats
WHERE id = $1;

-- name: UpdateCatPhoto :one
UPDATE cats SET
  photo_url = $1,
  updated_at = NOW()
WHERE id = $2
AND owner_id = $3
AND deleted_at IS NULL
RETURNING *;