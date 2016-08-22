# Set secret data

In secret.json file
```json
{
    "uid": "your user id",
    "client_id": "client id",
    "cookie": "your cookie"
}
```

In db.json file
```json
{
    "username": "your name",
    "password": "your password",
    "dbname": "database name"
}
```

# Start fetch user data

```sh
$ go run fetcher.go
```


# Start your server

```sh
$ go run server.go
```