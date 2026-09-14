.PHONY: run build test vet tidy seed db-create db-drop

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

# migrasi + seed RBAC & data demo ke PostgreSQL, lalu berhenti
seed:
	SEED_DEMO=true SEED_ONLY=true go run ./cmd/server

# membuat role & database PostgreSQL lokal
# (dijalankan dengan user postgres / superuser)
db-create:
	psql -d postgres -c "CREATE ROLE payroll LOGIN PASSWORD 'payroll';" || true
	createdb -O payroll payroll || true

db-drop:
	dropdb payroll || true
