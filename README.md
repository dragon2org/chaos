[![pipeline status](https://github.com/dragon2org/chaos/badges/master/pipeline.svg)](https://github.com/dragon2org/chaos/-/commits/master)
[![coverage report](https://github.com/dragon2org/chaos/badges/master/coverage.svg)](https://github.com/dragon2org/chaos/-/commits/master)
# jimo-argo-actor

chaos, you need a hammer

# HOWTO
```shell
$ go install github.com/dragon2org/chaos/chaos@latest

# person.go
//go:generate chaos clause -t Person -o person_clause.go

type Person struct {
  Name string `sql:"column:name"`
  LastName string // no sql tag, 不处理
  Age int `sql:"column:age"`
  
  gender string // not exportable, 不处理
}

$ go generate ./...
```