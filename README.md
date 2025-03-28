# Микросервисное Приложение для Управления Заказами Ресторана
### Описание
Cafe-services - микросервисное приложение, для управления заказами в ресторане. Оно состоит из четырех основных сервисов: приема заказов, аутентификации, кухни и доставки. Приложение использует различные технологии и инструменты для обеспечения эффективной работы и масштабируемости.

### Компоненты
- Auth Service Redis и Postgres
Хранит хешированные пароли пользователей.Генерирует и проверяет JWT токены. Отвечает за аутентификацию, регистрацию и авторизацию, сохраняет необходмую информацию в Redis и Postgres. 
- Order Service
Принимает новые заказы от клиентов. Общается с сервисом аутентификации с помощью фреймворка gRPC. Публикует события о новых заказах в Kafka.
Хранит заказы в кэше Redis для быстрого доступа.
- Kitchen Service
Подписывается на события о новых заказах из Kafka.
Обрабатывает заказы. Обновляет статус заказа и публикует события о готовности заказа.
- Delivery Service
Подписывается на события о готовности заказа из Kafka.
Обрабатывает доставку заказа. Сохраняет информацию о доставке в Redis и Postgres.

---

### Технологии и Инструменты
- **Язык:** Go 1.23.3
- **Брокер сообщений:** Apache Kafka
- **Кэширование:** Redis
- **База данных:** Postgres 
- **Docker-деплой:** Docker Compose
- **Прокси:** Caddy
- **Api:** gRPC, REST
- **Документация и протоколы:** 
  - **OpenAPI:** swaggo/swag
  - **AsyncAPI:** вручную в sayncapi.yaml
  - **Protobuf:** компилятор protocolbuffers/protobuf/releases, плагины protoc-gen-go, protoc-gen-go-grpc
- **CI/CD** GitHub Actions
  - .proto файл хранится в [proto-definitions](https://github.com/sandrinasava/proto-definitions) и при изменении автоматически обновляет 
  сгенерированный код в [go-proto-module](https://github.com/sandrinasava/go-proto-module)
