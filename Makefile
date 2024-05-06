build:
	docker-compose up -d --build
	@echo "Aplicativo compilado e iniciado com sucesso."

up:
	docker-compose up -d
	@echo "Aplicativo iniciado com sucesso."

down:
	docker-compose down
	@echo "Contêineres do aplicativo parados e removidos."

api:
	go run cmd/api/main.go
	@echo "Temperatures API iniciada com sucesso."

run:
	@curl -sv -H "Content-Type: application/json" http://localhost:8081/temperatures/89199000
	@echo # "Aplicativo executado com sucesso."

test:
	go test -v ./...
	@echo "Testes executados com sucesso."

logs:
	docker-compose logs app

.PHONY: go