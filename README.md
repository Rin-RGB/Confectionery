# SPOproject

Каркас backend-сервиса кондитерской.

Проект построен как модульный монолит с разделением HTTP-слоя, прикладных сервисов и репозиториев.
На текущем этапе создана только структура пакетов без бизнес-логики.

## Каталоги

- `backend/cmd/app` — точка входа приложения;
- `backend/internal/core` — доменные типы и общая инфраструктура;
- `backend/internal/handlers` — HTTP-обработчики и DTO;
- `backend/internal/service` — прикладные сценарии;
- `backend/internal/repository/postgres` — доступ к PostgreSQL;
- `backend/internal/transport` — адаптеры внешних сервисов;
- `backend/migrations` — SQL-миграции;
- `backend/docs` — OpenAPI-документация;
- `backend/tests/e2e` — сквозные тесты.
