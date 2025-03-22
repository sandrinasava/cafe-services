# Используем официальный образ PostgreSQL
FROM postgres:14

# Устанавливаем переменные окружения
ENV POSTGRES_USER=myuser
ENV POSTGRES_PASSWORD=mypassword
ENV POSTGRES_DB=mydatabase

# Копируем скрипты инициализации (если нужны)
# COPY init.sql /docker-entrypoint-initdb.d/