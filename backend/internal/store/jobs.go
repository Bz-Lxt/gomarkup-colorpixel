package store

import (
	"context"

	"colorpixel/internal/timeutil"
)

func (d *DB) EnqueueJob(ctx context.Context, assetID int64, kind string) (int64, error) {
	var id int64
	err := d.Pool.QueryRow(ctx, `
INSERT INTO jobs (asset_id, kind, status, attempts, created_at, updated_at)
VALUES ($1,$2,'queued',0,$3,$3) RETURNING id`, assetID, kind, timeutil.Now()).Scan(&id)
	return id, err
}

func (d *DB) ClaimJobs(ctx context.Context, limit int) ([]Job, error) {
	rows, err := d.Pool.Query(ctx, `
UPDATE jobs SET status='running', attempts=attempts+1, updated_at=$2
WHERE id IN (
  SELECT id FROM jobs WHERE status IN ('queued','failed') AND attempts < 5
  ORDER BY id ASC LIMIT $1 FOR UPDATE SKIP LOCKED
)
RETURNING id, asset_id, kind, status, attempts, COALESCE(last_error,'')`, limit, timeutil.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.AssetID, &j.Kind, &j.Status, &j.Attempts, &j.LastError); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (d *DB) RecoverJobs(ctx context.Context) error {
	_, err := d.Pool.Exec(ctx, `
UPDATE jobs SET status='queued', updated_at=$1
WHERE status='running'`, timeutil.Now())
	return err
}

func (d *DB) FinishJob(ctx context.Context, id int64, ok bool, lastErr string) error {
	st := "succeeded"
	if !ok {
		st = "failed"
	}
	_, err := d.Pool.Exec(ctx, `UPDATE jobs SET status=$2, last_error=$3, updated_at=$4 WHERE id=$1`,
		id, st, lastErr, timeutil.Now())
	return err
}

func (d *DB) JobStats(ctx context.Context) (JobStats, error) {
	var s JobStats
	err := d.Pool.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE status='queued'),
  COUNT(*) FILTER (WHERE status='running'),
  COUNT(*) FILTER (WHERE status='succeeded'),
  COUNT(*) FILTER (WHERE status='failed')
FROM jobs`).Scan(&s.Queued, &s.Running, &s.Succeeded, &s.Failed)
	return s, err
}
