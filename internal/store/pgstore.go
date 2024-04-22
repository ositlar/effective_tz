package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
)

type PGStore struct {
	db *sql.DB
}

func NewPGStore(db *sql.DB) *PGStore {
	return &PGStore{db: db}
}

func (s *PGStore) Create(num string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, "INSERT INTO numbers(number) VALUES($1)", num)
	if err != nil {
		return err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return errors.New("timeout")
	}
	return nil
}

func (s *PGStore) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, "DELETE FROM numbers WHERE id=$1", id)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return errors.New("timeout")
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *PGStore) GetById(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, "SELECT number FROM numbers WHERE id=$1", id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", errors.New("timeout")
	}
	var number string
	if rows.Next() {
		if err := rows.Scan(&number); err != nil {
			return "", err
		}
	}
	return number, nil
}

func (s *PGStore) GetByPrefix(prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pf := "%" + prefix + "%"
	rows, err := s.db.QueryContext(ctx, "SELECT number FROM numbers WHERE number LIKE $1", pf)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, errors.New("timeout")
	}
	var numbers []string
	for rows.Next() {
		var number string
		if err := rows.Scan(&number); err != nil {
			return nil, err
		}
		numbers = append(numbers, number)
	}
	//fmt.Println(len(numbers))
	return numbers, nil
}

func (s *PGStore) GetByRegion(region string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, "SELECT number FROM numbers WHERE number ~ $1", "[[:alpha:]].*"+region+"$")
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, errors.New("timeout")
	}
	var numbers []string
	for rows.Next() {
		var number string
		if err := rows.Scan(&number); err != nil {
			return nil, err
		}
		numbers = append(numbers, number)
	}
	//fmt.Println(len(numbers))
	return numbers, nil
}

func (s *PGStore) Update(id, newNum string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var result string
	_ = s.db.QueryRowContext(ctx, "UPDATE numbers SET number = $2 WHERE id = $1 RETURNING id", id, newNum).Scan(&result)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", errors.New("timeout")
	}
	if result == "" {
		return "", errors.New("no rows affected")
	}
	return id, nil
}

func (s *PGStore) CreateEnriched(data map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	regNum := data["regNum"].(string)
	mark := data["mark"].(string)
	model := data["model"].(string)
	year := data["year"].(int16)
	owner := data["owner"].(map[string]interface{})
	name := owner["name"].(string)
	surname := owner["surname"].(string)
	patronymic := owner["patronymic"].(string)
	_, err := s.db.ExecContext(ctx, "INSERT INTO enriched_info(regNum, mark, model ,year, name, surname, patronymic) VALUES($1, $2, $3, $4, $5, $6, $7)",
		regNum, mark, model, year, name, surname, patronymic)
	if err != nil {
		return err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return errors.New("timeout")
	}
	return nil
}

func (s *PGStore) Migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	queryNumbers := "CREATE TABLE IF NOT EXISTS numbers (id BIGSERIAL PRIMARY KEY, number text NOT NULL);"
	queryEnriched_info := "CREATE TABLE IF NOT EXISTS enriched_info (id BIGSERIAL PRIMARY KEY,regNum text NOT NULL,mark text NOT NULL,model text NOT NULL,year integer NOT NULL,name text NOT NULL,surname text NOT NULL,patronymic text);"
	_, err := s.db.ExecContext(ctx, queryNumbers)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, queryEnriched_info)
	if err != nil {
		return err
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return errors.New("timeout")
	}
	return nil
}
