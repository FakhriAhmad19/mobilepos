package providers

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"

	"goravel/database/seeders"
)

// DatabaseServiceProvider registers the application's seeders so they can be
// executed with `artisan db:seed`.
type DatabaseServiceProvider struct{}

func (r *DatabaseServiceProvider) Register(app foundation.Application) {}

func (r *DatabaseServiceProvider) Boot(app foundation.Application) {
	facades.Seeder().Register([]seeder.Seeder{
		&seeders.DatabaseSeeder{},
	})
}
