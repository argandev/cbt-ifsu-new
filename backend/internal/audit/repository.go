package audit

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket audit.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// strPtr mengubah string kosong menjadi NULL untuk parameter SQL.
// (Dipindah dari audit.go tanpa perubahan.)
func strPtr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// ListLog membaca baris audit_log sesuai filter. (Dulu: query inline di
// handler GET /api/v1/audit.)
func (r *Repository) ListLog(ctx context.Context, aksi, entitas, actorID string, limit, offset int) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, actor_id, aksi, entitas, entitas_id, detail, ip, created_at
		 FROM audit_log
		 WHERE ($1::text IS NULL OR aksi = $1)
		   AND ($2::text IS NULL OR entitas = $2)
		   AND ($3::uuid IS NULL OR actor_id = $3)
		 ORDER BY created_at DESC
		 LIMIT $4 OFFSET $5`,
		strPtr(aksi), strPtr(entitas), strPtr(actorID), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []map[string]any{}
	for rows.Next() {
		var id, aksiV, entitasV string
		var actorIDV, entitasIDV, ipV *string
		var detail json.RawMessage
		var createdAt any
		if err := rows.Scan(&id, &actorIDV, &aksiV, &entitasV, &entitasIDV, &detail, &ipV, &createdAt); err != nil {
			return nil, err
		}
		var detailAny any
		if len(detail) > 0 {
			_ = json.Unmarshal(detail, &detailAny)
		}
		items = append(items, map[string]any{
			"id": id, "actor_id": actorIDV, "aksi": aksiV, "entitas": entitasV,
			"entitas_id": entitasIDV, "detail": detailAny, "ip": ipV, "created_at": createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
