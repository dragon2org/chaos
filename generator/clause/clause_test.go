// nolint
//
//go:generate go run github.com/dragon2org/chaos/chaos clause -t=Employee -t=Square -output clauses/person_clause.go
package clause

import (
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"
)

type Employee struct {
	// base_test.BaseMode
	BaseMode

	Name      string       `sql:"column:name"`
	Age       string       `sql:"-"`
	Wage      string       `gorm:"column:wage"`
	Gender    string       `sql:"gender;type:varchar(2)"`
	Active    bool         `db:"active"`
	DisableAt sql.NullTime `gorm:"column:disable_at"`
	Birthday  time.Time    `gorm:"column:birthday"`
	JoinAt    *time.Time   `gorm:"column:join_at"`
	Family    StringSlice  `gorm:"column:family"`
}

type Square struct {
	Width  int `gorm:"column:width"`
	Height int `gorm:"column:height"`
}

type StringSlice []string

func (s *StringSlice) Value() (driver.Value, error) {
	panic("implement me")
}

func TestGeneratorV2_Execute(t *testing.T) {
	gen := NewGenerator([]string{"Employee", "Square"}, "", "")
	gen.Execute()
}
