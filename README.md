# Simple Migrations for `jackc/pgx`

Very light weight and opinionated migration framework. It only supports "up" migrations, PostgreSQL database, golang app, `jackc/pgx` database client, `int64` migration keys. You get to pick the table and column name. 

Shell command to get a new migration key:

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
  
  // etc

  err = m.Exec(conn)
  if err != nil {
    log.Fatal(err)
  }
}
```
