-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS cats (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  owner_id UUID NOT NULL,
  name VARCHAR(100) NOT NULL,
  breed VARCHAR(100),
  date_of_birth DATE,
  weight DECIMAL(5, 2), -- in kg
  color VARCHAR(50),
  gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'unknown')),
  photo_url TEXT,
  medical_notes TEXT,
  dietary_requirements TEXT,
  date_of_death DATE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_cats_owner_id ON cats(owner_id);
CREATE INDEX IF NOT EXISTS idx_cats_alive ON cats(date_of_death)
WHERE date_of_death IS NULL;
CREATE INDEX IF NOT EXISTS idx_cats_active ON cats(deleted_at)
WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_cats_owner_id;
DROP INDEX IF EXISTS idx_cats_alive;
DROP INDEX IF EXISTS idx_cats_active;
DROP TABLE IF EXISTS cats;
-- +goose StatementEnd
