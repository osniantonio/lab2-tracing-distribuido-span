build:
	docker-compose up -d --build
	@echo "Aplicativo compilado e iniciado com sucesso."

up:
	docker-compose up -d
	@echo "Aplicativo iniciado com sucesso."

down:
	docker-compose down
	@echo "Contêineres do aplicativo parados e removidos."

logs:
	docker-compose logs serviceb

.PHONY: go