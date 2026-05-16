#скрыть команды
.SILENT:

.PHONY: up down restart test logs

#поднять в фоне
up:
	echo "Запуск контейнеров..."
	docker compose up --build -d

#остановка контейнеров
down:
	echo "Контейрены останавливаются.."
	docker compose down

#ребут
restart: down up

#логи
logs:
	docker compose logs -f api

#тесты
test:
	go test ./... -v