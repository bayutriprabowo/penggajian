.PHONY: run build test vet tidy db-create db-drop

run:
	go run ./cmd/server

build:
	go build -o bin/penggajian ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

# membuat role & database PostgreSQL lokal
# (dijalankan dengan user postgres / superuser)
db-create:
	psql -d postgres -c "CREATE ROLE payroll LOGIN PASSWORD 'payroll';" || true
	createdb -O payroll payroll || true

db-drop:
	dropdb payroll || true
