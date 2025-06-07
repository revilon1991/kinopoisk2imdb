# Консольное приложение для переноса фильмов с КиноПоиска на IMDb

Это консольное приложение на Go поможет поэтапно перенести списки фильмов с сайта [КиноПоиск](https://www.kinopoisk.ru) на сайт [IMDb](https://www.imdb.com).

## ⚙️ Требования

- Go 1.24+
- Поверхностные знания Go, чтобы запускать нужные этапы
- Cookie-данные из аккаунтов КиноПоиска и IMDb

## 📁 Подготовка

Перед использованием нужно заполнить следующие **константы** в `main.go` своими значениями:

| Константа            | Описание                                        |
|----------------------|-------------------------------------------------|
| `readCsv`            | Путь к CSV-файлу, из которого читаются данные   |
| `writeCsv`           | Путь к CSV-файлу, в который записываются данные |
| `KpCookieYaSessId`   | Cookie `ya-sess-id` из КиноПоиска               |
| `KpWatchListId`      | ID списка «Найти в интернете»                   |
| `KpFavoriteListId`   | ID списка «Любимые фильмы»                      |
| `ImdbCookieAtMain`   | Cookie `at-main` из IMDb                        |
| `ImdbCookieUbidMain` | Cookie `ubid-main` из IMDb                      |
| `ImdbCheckInListId`  | ID списка check-in на IMDb                      |
| `ImdbFavoriteListId` | ID списка любимых фильмов на IMDb               |

## 🧠 Этапы работы

Каждая функция запускается вручную, раскомментировав её в `main()`.

### 1. Сбор фильмов с КиноПоиска

- `kpListParser()` — парсит названия, годы и оценки из списков «Найти в интернете» и «Любимые».
- `kpWatchedListParser()` — парсит из HTML-страниц в папке `pages` (1.html, 2.html и т.д.)

### 2. Маппинг на IMDb

- `imdbMapping()` — ищет соответствия фильмов в IMDb и записывает в `writeCsv`. Точность ~95%. Требуется ручная проверка.
- `showDuplicates()` — показывает дубликаты IMDb ID в файле для дополнительной проверки.

### 3. Добавление фильмов на IMDb

- `addFilmToImdbList(listId string, startRow int)` — добавляет фильмы в указанный список IMDb.
- `addFilmToImdbWatchList(startRow int)` — добавляет в Watchlist.
- `rateFilmToImdb()` — выставляет оценки на IMDb (если были в КиноПоиске).

## ⚠️ Замечания

- Точность маппинга — примерно 95%.
- После `imdbMapping()` рекомендуется вручную проверить и исправить ошибки в `writeCsv`.

## 📦 Запуск

1. Установите зависимости:
   ```shell
   go mod tidy
   ```
2.	Запустите нужный этап, раскомментировав соответствующую функцию в main().
3.	Выполните:
   ```shell
   go run main.go
   ```

License
-------

[![license](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](./LICENSE)
