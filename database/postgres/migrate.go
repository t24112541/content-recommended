package postgres

import (
	"content-recommended/model/orm"
)

func (r *resourceDB) migrate() (err error) {

	err = r.db.AutoMigrate(

		// ?: add model here
		&orm.Users{},
		&orm.Content{},
		&orm.UserWatchHistory{},
	)

	return
}
