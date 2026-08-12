package DBMigration

import "testing"

func TestLoadMigrations(t *testing.T) {
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("loadMigrations() returned no migrations")
	}
	for i, m := range migrations {
		if m.SQL == "" {
			t.Fatalf("migration %d_%s has empty SQL", m.Version, m.Name)
		}
		if i > 0 && migrations[i-1].Version >= m.Version {
			t.Fatalf("migrations not sorted ascending at %d_%s", m.Version, m.Name)
		}
	}
	if migrations[0].Version != 1 {
		t.Fatalf("first migration version = %d, want 1", migrations[0].Version)
	}
}

func TestMigrationFilePattern(t *testing.T) {
	cases := []struct {
		name    string
		version string
		ok      bool
	}{
		{"0001_baseline.sql", "0001", true},
		{"0002_add_xxx.sql", "0002", true},
		{"00000001_ok.sql", "00000001", true},
		{"0001.sql", "", false},
		{"baseline.sql", "", false},
		{"0001_has_extra.txt", "", false},
	}
	for _, c := range cases {
		match := migrationFilePattern.FindStringSubmatch(c.name)
		if c.ok {
			if match == nil || match[1] != c.version {
				t.Errorf("pattern mismatch for %q: got %v", c.name, match)
			}
		} else if match != nil {
			t.Errorf("expected %q to be rejected, got %v", c.name, match)
		}
	}
}
