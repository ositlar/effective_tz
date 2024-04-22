package store

type Store interface {
	Migrate() error
	Create(string) error
	Delete(string) error
	GetById(string) (string, error)
	GetByPrefix(string) ([]string, error)
	GetByRegion(string) ([]string, error)
	Update(string, string) (string, error)
	CreateEnriched(map[string]interface{}) error
}
