# Migration library for Golang apps

Light-weight migration framework for golang apps. It only supports "up" migrations and only works with databases that support transactions. It only works with `int64` migration keys. 

Create a new timestamp migration key:

```sh
$ date +%s
```

## Example

```go
package main

import (
  "context"
  "github.com/aj0strow/migrate/jackc-pgx-migrate"
  "github.com/jackc/pgx"
  "log"
  "os"
)

func migrate() {
  conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
  if err != nil {
    log.Fatal(err)
  }
  defer conn.Close()
  
  m := migrate.New()

  m.Set(1575346891, func(tx *pgx.Tx) error {
    migrate.Run(tx, []string{
      `CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        email TEXT NOT NULL
      )`,
      `CREATE UNIQUE INDEX users_email_key ON users (email)`,
    })
  })

  if  err := m.Exec(conn); err != nil {
    log.Fatal(err)
  }
}
```

I personally keep more recent migrations at the top of the file and periodically flatten migrations into the minimum possible set. It's important when flattening to never add new migrations and only ever remove existing migrations that already ran in all environments. 

```go
  m.Set(1575348887, func (tx *pgx.Tx) error {
    migrate.Run(tx, []string{
      `ALTER TABLE users ADD COLUMN name TEXT NOT NULL DEFAULT ''`,
    })
  })

  m.Set(1575346891, func(tx *pgx.Tx) error {
    migrate.Run(tx, []string{
      `CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        email TEXT NOT NULL
      )`,
      `CREATE UNIQUE INDEX users_email_key ON users (email)`,
    })
  })
  
  // Later becomes ..
  
  m.Set(1575348887, func (tx *pgx.Tx) error {
    migrate.Run(tx, []string{
      `CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        email TEXT NOT NULL,
        name TEXT NOT NULL DEFAULT ''
      )`,
      `CREATE UNIQUE INDEX users_email_key ON users (email)`,
    })
  })
```