package gomigrations

import (
	"github.com/TheAmirhosssein/goose/v3"
)

func init() {
	goose.AddMigrationNoTxContext(nil, nil)
}
