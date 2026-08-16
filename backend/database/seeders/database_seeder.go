package seeders

// DatabaseSeeder is the root seeder run by `artisan db:seed`. It invokes the
// individual seeders in dependency order (roles → users → master data → products).
type DatabaseSeeder struct{}

func (s *DatabaseSeeder) Signature() string {
	return "DatabaseSeeder"
}

func (s *DatabaseSeeder) Run() error {
	steps := []interface{ Run() error }{
		&RoleSeeder{},
		&UserSeeder{},
		&MasterSeeder{},
		&ProductSeeder{},
	}

	for _, step := range steps {
		if err := step.Run(); err != nil {
			return err
		}
	}

	return nil
}
