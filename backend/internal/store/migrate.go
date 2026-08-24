package store

import (
	"context"
	"fmt"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  username text UNIQUE NOT NULL,
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS assets (
  id bigserial PRIMARY KEY,
  filename text NOT NULL,
  format text NOT NULL,
  size_bytes bigint NOT NULL,
  storage_path text NOT NULL,
  preview_path text,
  extraction_mode text NOT NULL,
  camera_make text,
  camera_model text,
  lens_model text,
  lens_spec text,
  aperture double precision,
  shutter_text text,
  shutter_seconds double precision,
  iso integer,
  focal_length double precision,
  focal_length_35mm double precision,
  datetime_original timestamptz,
  orientation integer,
  white_balance text,
  exposure_bias double precision,
  rating integer NOT NULL DEFAULT 0,
  tags text[] NOT NULL DEFAULT '{}',
  sharpness double precision,
  noise double precision,
  clip_shadow double precision,
  clip_highlight double precision,
  ev_deviation double precision,
  tile_status text NOT NULL DEFAULT 'pending',
  tile_max_z integer NOT NULL DEFAULT 0,
  width integer,
  height integer,
  exif_raw jsonb NOT NULL DEFAULT '{}',
  deleted_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_assets_alive ON assets (datetime_original DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_assets_camera ON assets (camera_make, camera_model) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_assets_lens ON assets (lens_model) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_assets_exif ON assets USING gin (exif_raw);
CREATE INDEX IF NOT EXISTS idx_assets_iso ON assets (iso) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS jobs (
  id bigserial PRIMARY KEY,
  asset_id bigint NOT NULL REFERENCES assets(id),
  kind text NOT NULL,
  status text NOT NULL,
  attempts integer NOT NULL DEFAULT 0,
  last_error text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs (status, id);
`

func (d *DB) Migrate(ctx context.Context) error {
	conn, err := d.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_lock(884421)`); err != nil {
		return fmt.Errorf("advisory lock: %w", err)
	}
	defer tx.Exec(ctx, `SELECT pg_advisory_unlock(884421)`)
	if _, err := tx.Exec(ctx, schema); err != nil {
		return fmt.Errorf("schema: %w", err)
	}
	return tx.Commit(ctx)
}
