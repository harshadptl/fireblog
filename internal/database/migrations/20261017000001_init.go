package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	// Snapshot of the initial schema. Works on both PostgreSQL and SQLite.
	type user struct {
		bun.BaseModel `bun:"table:users"`
		ID            int64     `bun:",pk,autoincrement"`
		GoogleID      string    `bun:",unique,notnull"`
		Email         string    `bun:",unique,notnull"`
		Name          string    `bun:",notnull,default:''"`
		Username      string    `bun:",unique,notnull"`
		AvatarURL     string    `bun:",notnull,default:''"`
		Bio           string    `bun:",type:text,notnull,default:''"`
		Role          string    `bun:",notnull,default:'user'"`
		LastLoginAt   time.Time `bun:",nullzero"`
		CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		UpdatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	}
	type post struct {
		bun.BaseModel  `bun:"table:posts"`
		ID             int64     `bun:",pk,autoincrement"`
		AuthorID       int64     `bun:",notnull"`
		Title          string    `bun:",notnull"`
		Slug           string    `bun:",unique,notnull"`
		Excerpt        string    `bun:",type:text,notnull,default:''"`
		Content        string    `bun:",type:text,notnull,default:''"`
		ContentHTML    string    `bun:"content_html,type:text,notnull,default:''"`
		CoverImage     string    `bun:",notnull,default:''"`
		Status         string    `bun:",notnull,default:'draft'"`
		ReadingMinutes int       `bun:",notnull,default:1"`
		Views          int64     `bun:",notnull,default:0"`
		PublishedAt    time.Time `bun:",nullzero"`
		CreatedAt      time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		UpdatedAt      time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	}
	type tag struct {
		bun.BaseModel `bun:"table:tags"`
		ID            int64     `bun:",pk,autoincrement"`
		Name          string    `bun:",notnull"`
		Slug          string    `bun:",unique,notnull"`
		CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	}
	type postTag struct {
		bun.BaseModel `bun:"table:post_tags"`
		PostID        int64 `bun:",pk"`
		TagID         int64 `bun:",pk"`
	}
	type comment struct {
		bun.BaseModel `bun:"table:comments"`
		ID            int64     `bun:",pk,autoincrement"`
		PostID        int64     `bun:",notnull"`
		AuthorID      int64     `bun:",notnull"`
		Body          string    `bun:",type:text,notnull"`
		CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		UpdatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	}
	type like struct {
		bun.BaseModel `bun:"table:likes"`
		UserID        int64     `bun:",pk"`
		PostID        int64     `bun:",pk"`
		CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	}
	type bookmark struct {
		bun.BaseModel `bun:"table:bookmarks"`
		UserID        int64     `bun:",pk"`
		PostID        int64     `bun:",pk"`
		CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	}
	type session struct {
		bun.BaseModel `bun:"table:sessions"`
		ID            string    `bun:",pk"`
		UserID        int64     `bun:",notnull"`
		CSRFToken     string    `bun:"csrf_token,notnull"`
		UserAgent     string    `bun:",notnull,default:''"`
		IP            string    `bun:"ip,notnull,default:''"`
		CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		LastSeenAt    time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		ExpiresAt     time.Time `bun:",notnull"`
	}

	const cascade = "ON DELETE CASCADE"

	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			tables := []struct {
				model any
				fks   []string
			}{
				{(*user)(nil), nil},
				{(*post)(nil), []string{`("author_id") REFERENCES "users" ("id") ` + cascade}},
				{(*tag)(nil), nil},
				{(*postTag)(nil), []string{
					`("post_id") REFERENCES "posts" ("id") ` + cascade,
					`("tag_id") REFERENCES "tags" ("id") ` + cascade,
				}},
				{(*comment)(nil), []string{
					`("post_id") REFERENCES "posts" ("id") ` + cascade,
					`("author_id") REFERENCES "users" ("id") ` + cascade,
				}},
				{(*like)(nil), []string{
					`("user_id") REFERENCES "users" ("id") ` + cascade,
					`("post_id") REFERENCES "posts" ("id") ` + cascade,
				}},
				{(*bookmark)(nil), []string{
					`("user_id") REFERENCES "users" ("id") ` + cascade,
					`("post_id") REFERENCES "posts" ("id") ` + cascade,
				}},
				{(*session)(nil), []string{`("user_id") REFERENCES "users" ("id") ` + cascade}},
			}
			for _, t := range tables {
				q := tx.NewCreateTable().Model(t.model)
				for _, fk := range t.fks {
					q = q.ForeignKey(fk)
				}
				if _, err := q.Exec(ctx); err != nil {
					return err
				}
			}

			indexes := []struct {
				model   any
				name    string
				columns []string
			}{
				{(*post)(nil), "posts_status_published_at_idx", []string{"status", "published_at"}},
				{(*post)(nil), "posts_author_id_idx", []string{"author_id"}},
				{(*postTag)(nil), "post_tags_tag_id_idx", []string{"tag_id"}},
				{(*comment)(nil), "comments_post_id_idx", []string{"post_id", "created_at"}},
				{(*like)(nil), "likes_post_id_idx", []string{"post_id"}},
				{(*bookmark)(nil), "bookmarks_post_id_idx", []string{"post_id"}},
				{(*session)(nil), "sessions_user_id_idx", []string{"user_id"}},
				{(*session)(nil), "sessions_expires_at_idx", []string{"expires_at"}},
			}
			for _, ix := range indexes {
				if _, err := tx.NewCreateIndex().Model(ix.model).Index(ix.name).Column(ix.columns...).Exec(ctx); err != nil {
					return fmt.Errorf("create index %s: %w", ix.name, err)
				}
			}
			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			for _, m := range []any{
				(*session)(nil), (*bookmark)(nil), (*like)(nil), (*comment)(nil),
				(*postTag)(nil), (*tag)(nil), (*post)(nil), (*user)(nil),
			} {
				if _, err := tx.NewDropTable().Model(m).IfExists().Exec(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	}

	Migrations.MustRegister(up, down)
}
