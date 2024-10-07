Реализация Онлайн Библиотеки песен

Swagger UI открывается по адресу: http://localhost:8080/swagger/index.html#/
Запросы:
POST /songs - добавить песню, обязательные параметры: group, song, необязательные: text, link, releaseDate
PUT /songs - редактировать песню, обязательные параметры: group, song, необязательные: text, link, releaseDate
DELETE /songs - удалить песню, обязательные параметры: group, song
GET /songs/text - получить текст песни с пагинацией, обязательные параметры: group, song
GET /info - получить releaseDate, text, link, указанной песни, обязательные парметры: group, song
GET /songs - получить список спесен с фильтрацией и пагинацией
