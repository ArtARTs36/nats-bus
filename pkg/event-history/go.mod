module github.com/artarts36/nats-bus/pkg/consumer-history

go 1.21.0

replace (
	github.com/artarts36/nats-bus => ./../..
)

require (
	github.com/artarts36/nats-bus v0.1.1
	github.com/doug-martin/goqu/v9 v9.19.0
	github.com/jmoiron/sqlx v1.4.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/nats-io/nats.go v1.34.1 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)