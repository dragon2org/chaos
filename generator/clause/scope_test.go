//go:generate go run github.com/dragon2org/chaos/command scope -t=Person -t=Car -output scopes/person_scope.go

package clause

import (
	"database/sql"
	"testing"
	"time"
)

type BaseMode struct {
	ID        string       `sql:"primary_key;column:id"`
	CreatedAt time.Time    `gorm:"column:created_at"`
	DeletedAt sql.NullTime `gorm:"column:deleted_at"`
}

type Person struct {
	BaseMode
	Name   string `sql:"column:name"`
	Gender string `sql:"-"`
	Age    int    `gorm:"column:age"`
}

type Car struct {
	Engine string `sql:"column:engine"`
}

func TestGen(t *testing.T) {
	Gen([]string{"Person", "Car"}, "")
}
