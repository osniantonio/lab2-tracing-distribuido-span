build:
	docker-compose up -d --build
	@echo "Aplicativo compilado e iniciado com sucesso."

up:
	docker-compose up -d
	@echo "Aplicativo iniciado com sucesso."

down:
	docker-compose down
	@echo "Contêineres do aplicativo parados e removidos."

test:
	go test -v ./...
	@echo "Testes executados com sucesso."

logs:
	docker-compose logs serviceb

.PHONY: go