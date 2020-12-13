GOOS?=linux

gen:
	protoc \
	--proto_path=services/proto/source \
	--go_out=services/proto/build/products \
	--go_opt=paths=source_relative \
	--go-grpc_out=services/proto/build/products \
	--go-grpc_opt=paths=source_relative \
	services/proto/source/products.proto


build:
	go mod tidy
	GOOS=$(GOOS) go build -o .build/prod-catalog cmd/prod_srv/main.go
	GOOS=$(GOOS) go build -o .build/csv-source cmd/csv_source/main.go

run: build
	docker-compose up --build -d

stop:
	docker-compose down

test: run
	go test -v ./...
	go test -bench=. docker_compose_test.go
	docker-compose down