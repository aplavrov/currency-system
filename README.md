# Currency Wallet System

Система из трех сервисов:
- кошелек-обменник с авторизацией (**gw-currency-wallet**)
- сервис exchanger для получения курсов валют (**gw-exchanger**)
- сервис notification для получения больших денежных переводов (**gw-notification**)

## Запуск приложения

```bash
docker-compose up --build
````

Будут запущены:

* postgres — основная база данных (миграции применяются в коде), используется для хранения курсов валют в **gw-exchanger**, а также для хранения пользователей и их кошельков в **gw-currency-wallet** 
* kafka — используется для передачи сообщений о больших транзакциях от сервиса **gw-currency-wallet** в сервис **gw-notification**
* mongodb — используется для хранения полученных сообщений о больших транзакциях в сервисе **gw-notification**
* gw-currency-wallet, gw-exchanger, gw-notification — сервисы

Работа с системой устроена через REST с сервисом **gw-currency-wallet** по адресу:

```
http://localhost:9000
```

На нем запущен Swagger, доступный по адресу:

```
http://localhost:9000/swagger/index.html
```

## gw-currency-wallet
Доступны следующие методы:
* Регистрация пользователя **POST** **/api/v1/register** 
* Авторизация пользователя **POST** **/api/v1/login**
* Получение баланса пользователя **GET** **/api/v1/balance**
* Пополнение счета **POST** **/api/v1/wallet/deposit**
* Вывод средств **POST** **/api/v1/wallet/withdraw**
* Получение курса валют **GET** **/api/v1/exchange/rates**
* Обмен валют **POST** **/api/v1/exchange**

Все запросы защищены JWT-токенами. Запросы курсов валют сохраняются в in-memory cache.

## gw-exchanger
Сервис, реализующий gRPC-сервер для получения курсов валют, к которому обращается **gw-currency-wallet**.

## gw-notification
Сервис, читающий сообщения из Kafka о больших транзакциях ($\geq$ 30000) и сохраняющий информацию о них в MongoDB.