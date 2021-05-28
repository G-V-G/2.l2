# 2 / l2,3

![travis test](https://travis-ci.org/G-V-G/2.l2.svg?branch=master)

### Запуск тестування
> docker-compose -f docker-compose.yaml -f docker-compose.test.yaml up --exit-code-from test

### Запуск серверів та балансувальника
> docker-compose up

### Запуск клієнта (для перевірки відправки запитів)
> go run ./cmd/client
