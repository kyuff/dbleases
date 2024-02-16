

db-reset:
	echo "drop schema public cascade" 			| docker exec -i leasedb psql -U lease lease
	echo "create schema if not exists public" 	| docker exec -i leasedb psql -U lease lease


reset: db-reset

test:
	go test ./... -count 1 -race

test-it: up
	cd tests
	go test ./... -count 1 -race

test-all: test test-it

up:
	docker-compose up -d

down:
	docker-compose down
