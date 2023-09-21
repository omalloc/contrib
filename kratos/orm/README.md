## Kratos ORM

Kratos ORM is a GORM for Kratos Framework.

- Support OpenTelemetry (`Statement.SkipHooks` skip the reporting if not recording)

## Installation

### with cgo

[Docs](https://github.com/go-gorm/sqlite)

```shell
$ go get -u gorm.io/driver/sqlite
```

### without cgo

[Docs](https://github.com/glebarez/sqlite)

```shell
$ go get -u github.com/glebarez/sqlite
```

## Usage

### crud

```go
type MyModel struct {
    ID int64
    Name string

    crud.DBModel
}

// or interface.
type myRepo struct {
    crud.CRUD[MyModel]
}

func NewMyRepo(db *gorm.DB) *myRepo {
    return &myRepo{crud.New(db)}
}
```
