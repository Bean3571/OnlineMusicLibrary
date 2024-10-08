Реализация Онлайн Библиотеки песен

Для запуска, находясь в папке проекта, ввести в терминал: go run .

Swagger UI открывается по адресу: http://localhost:8080/swagger/index.html#/

В .env задаются данные для подключения к БД

БД создается путём миграций

Для логирования использован slog с tint slog.Handler, он позволяет использовать цвета при выводе логов

Запросы:

POST /songs - добавить песню, обязательные параметры: group, song, необязательные: text, link, releaseDate

PUT /songs - редактировать песню, обязательные параметры: group, song, необязательные: text, link, releaseDate

DELETE /songs - удалить песню, обязательные параметры: group, song

GET /songs/text - получить текст песни с пагинацией, обязательные параметры: group, song

GET /info - получить releaseDate, text, link, указанной песни, обязательные парметры: group, song

GET /songs - получить список спесен с фильтрацией и пагинацией

Swagger:
![{F9ED3FCD-4063-4676-9469-B977C9420C8B}](https://github.com/user-attachments/assets/78163bd4-5802-41ea-bb50-7ff13e04ba75)

Отрывок из лога:
![{DAE22A2C-EDB2-46AE-8603-E9BB898FD4CC}](https://github.com/user-attachments/assets/718c50fb-df9f-4740-acfa-0c7859b1ec8c)

