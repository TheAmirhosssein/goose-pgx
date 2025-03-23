package register

import (
	"database/sql"

	"github.com/TheAmirhosssein/goose/v3"
)

func init() {
	goose.AddMigration(
		func(_ *sql.Tx) error { return nil },
		func(_ *sql.Tx) error { return nil },
	)
}
