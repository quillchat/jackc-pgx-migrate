# Simple Migrations for `jackc/pgx`

Quick and opinionated migration framework. It only supports "up" migrations, PostgreSQL database, golang app, `jackc/pgx` database client, `int64` migration keys. It will create a new database table `migrations` with primary key `mts` which stands for *migration unix timestamp*. 

Shell command to create a new migration key:

```sh
$ date +%s
```

Go import path:

```
github.com/aj0strow/jackc-pgx-migrate
```

## Example

```go
package main

import (
  "context"
  "github.com/aj0strow/jackc-pgx-migrate"
  "github.com/jackc/pgx/v4"
  "log"
  "os"
)

func migrate() {
  conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
  if err != nil {
    log.Fatal(err)
  }
  defer conn.Close()
  
  m := migrate.Funcs{}

  m[1575346891] = func(ctx context.Context, tx pgx.Tx) error {
    return migrate.Run(ctx, tx, []string{
      `CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        email TEXT NOT NULL
      )`,
      `CREATE UNIQUE INDEX users_email_key ON users (email)`,
    })
  }

  // etc

  err = migrate.Migrate(ctx, conn, m)
  if err != nil {
    log.Fatal(err)
  }
}
```

## Simple Benefits

Complex migration updating lots of data:

```go
  m[1577567921] func(ctx context.Context, tx pgx.Tx) error {
    // do whatever you want
    store := &Store{Tx: tx}
  }
```

Simple migration running commands:

```go
  m[1577567921] func(ctx context.Context, tx pgx.Tx) error {
    return migrate.Run(ctx, tx, []string{
      `CREATE TABLE ...`,
      `CREATE INDEX ...`,
    })
  }
```

Adding migrations from different parts of the app:

```go
package subapp

import (
  "github.com/aj0strow/jackc-pgx-migrate"
)

func MigrateFuncs() migrate.Funcs {
  m := migrate.Funcs{}

  // add migration functions

  return m
}
```

```go
package main

import (
  "github.com/aj0strow/jackc-pgx-migrate"
  "subapp"
)

func migrate() {
  m := migrate.Funcs{}
  for k, v := range subapp.MigrateFuncs() {
    m[k] = v
  }
  // add migration funcions
}
```

## Running Tests

Create a new database:

```sh
$ createdb jackc_pgx_migrate
```

Set the database connection string before running tests:

```sh
DATABASE_URL=postgres://localhost/jackc_pgx_migrate go test .
```

It sets the `search_path` to `'pg_temp'` to avoid cleanup between test runs. 
