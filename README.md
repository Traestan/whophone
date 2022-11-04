# Whophone
Сервис который собирает данные с opendata.digital.gov.ru но номерной вместимости. Хранит все в sqlite3.

## Конфигурирование
``` toml
[db]
  bucket="whophone" - название основной таблицы
  path = "whophone.db" - имя файла для sqlite

[numcap]
  filename=[ - имя файлов для номерной сетки 
    "ABC-3xx.csv",
	  "ABC-4xx.csv",
	  "ABC-8xx.csv",
	  "DEF-9xx.csv",
  ]
  pathfilestore="../data" - путь файлов для скачивания
```
## Сборка
Проверялось под linux
```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o whophone ./cmd/whophone
```

## Обновление базы
Так как номерная сетка может переходить от оператора к оператору то данные очищаются и вливаются с нуля.
```
whophone -action=update
```

## Поиск в базе
Поиск происходит так 

```
whophone -action=find -phone=89544256777
```

