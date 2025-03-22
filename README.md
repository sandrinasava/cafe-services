# Микросервисное Приложение для Управления Заказами Ресторана
### Описание
Этот проект представляет собой микросервисное приложение для управления заказами в ресторане. Оно состоит из трех основных сервисов: приема заказов, кухни и доставки. Приложение использует различные технологии и инструменты для обеспечения эффективной работы и масштабируемости.

### Компоненты
- Order Service
Принимает новые заказы от клиентов.
Публикует события о новых заказах в Kafka.
Хранит заказы в кэше Redis для быстрого доступа.
- Kitchen Service
Подписывается на события о новых заказах из Kafka.
Обрабатывает заказы (например, готовит пиццу).
Обновляет статус заказа и публикует события о готовности заказа.
- Delivery Service
Подписывается на события о готовности заказа из Kafka.
Обрабатывает доставку заказа.
Сохраняет информацию о доставке в базу данных PostgreSQL.

---

### Технологии и Инструменты
- **Язык:** Go 1.24
- **Брокер сообщений:** Apache Kafka
- **Кэширование:** Redis
- **База данных:** PostgreSQL 
- **Docker-деплой:** Docker Compose

---

### Запуск проекта

Order Service
Функциональность:

Принимает запросы на создание заказа.
Отправляет сообщение в Kafka для Cooking Service.
Технологии:

gRPC для взаимодействия с Auth Service.
Kafka для отправки сообщений в Cooking Service.
Auth Service
Функциональность:

Хранит хешированные пароли пользователей.
Генерирует и проверяет JWT токены.
Технологии:

gRPC для взаимодействия с другими микросервисами.
База данных (например, SQLite) для хранения пользователей.
Cooking Service
Функциональность:

Имитирует процесс приготовления заказа с помощью таймаута.
Отправляет сообщение в Kafka для Delivery Service после завершения приготовления.
Технологии:

Kafka для получения сообщений от Order Service и отправки сообщений в Delivery Service.
Delivery Service
Функциональность:

Имитирует процесс доставки заказа с помощью таймаута.
Обновляет статус заказа в базе данных.
Технологии:

Kafka для получения сообщений от Cooking Service.
База данных (например, SQLite) для хранения статуса заказов.
3. Взаимодействие микросервисов
gRPC:

Используйте Protocol Buffers для определения сообщений и сервисов.
Реализуйте клиенты и серверы gRPC для взаимодействия между микросервисами.
Kafka:

Настройте топики для каждого типа сообщений (например, order_created, order_cooked, order_delivered).
Используйте библиотеку Kafka для Go (например, sarama) для отправки и получения сообщений.
