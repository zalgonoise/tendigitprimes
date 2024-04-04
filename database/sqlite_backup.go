package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"modernc.org/sqlite"
)

type backuper interface {
	NewBackup(dstUri string) (*sqlite.Backup, error)
	NewRestore(srcUri string) (*sqlite.Backup, error)
}

func BackupTo(ctx context.Context, db *sql.DB, uri string, logger *slog.Logger) error {
	srcConn, err := db.Conn(ctx)
	if err != nil {
		return err
	}

	if err := srcConn.Raw(func(srcConnRaw any) error {
		s, ok := (srcConnRaw).(backuper)
		if !ok {
			return fmt.Errorf("invalid conn type: %T", srcConnRaw)
		}

		b, err := s.NewBackup(fmt.Sprintf(uriFormat, uri))
		if err != nil {
			return err
		}

		if _, err = b.Step(-1); err != nil {
			return err
		}

		return b.Finish()
	}); err != nil {
		return err
	}

	return nil
}
