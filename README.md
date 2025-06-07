# CLI Tool for Transferring Movies from Kinopoisk to IMDb

This is a Go-based CLI tool designed to help you transfer your movie lists from [Kinopoisk](https://www.kinopoisk.ru) to [IMDb](https://www.imdb.com) step by step.

## ⚙️ Requirements

- Go 1.24+
- Basic knowledge of Go to run specific stages
- Cookie data from Kinopoisk and IMDb accounts

## 📁 Preparation

Before using the tool, fill in the following **constants** in `main.go`:

| Constant             | Description                        |
|----------------------|------------------------------------|
| `readCsv`            | Path to the CSV file to read from  |
| `writeCsv`           | Path to the CSV file to write to   |
| `KpCookieYaSessId`   | `ya-sess-id` cookie from Kinopoisk |
| `KpWatchListId`      | ID of the "Find on Internet" list  |
| `KpFavoriteListId`   | ID of the "Favorites" list         |
| `ImdbCookieAtMain`   | `at-main` cookie from IMDb         |
| `ImdbCookieUbidMain` | `ubid-main` cookie from IMDb       |
| `ImdbCheckInListId`  | Check-in list ID on IMDb           |
| `ImdbFavoriteListId` | Favorites list ID on IMDb          |

## 🧠 Workflow

Each stage is triggered manually by uncommenting the corresponding function in `main()`.

### 1. Collecting Movies from Kinopoisk

- `kpListParser()` — parses titles, years, and ratings from "Find on Internet" and "Favorites" lists.
- `kpWatchedListParser()` — parses from HTML files in the `pages` folder (named 1.html, 2.html, etc.).

### 2. Mapping to IMDb

- `imdbMapping()` — maps movies to IMDb by searching titles. Output is written to `writeCsv`. Approx. 95% accuracy. Manual verification recommended.
- `showDuplicates()` — finds and prints duplicate IMDb IDs for verification.

### 3. Adding Movies to IMDb

- `addFilmToImdbList(listId string, startRow int)` — adds movies to a given IMDb list.
- `addFilmToImdbWatchList(startRow int)` — adds to the Watchlist.
- `rateFilmToImdb()` — rates movies on IMDb if ratings are available from Kinopoisk.

## ⚠️ Notes

- Mapping accuracy is ~95%.
- Manual review of `writeCsv` is advised after mapping.

## 📦 Run

1. Install dependencies:
   ```shell
   go mod tidy
   ```
2. Uncomment the desired function in main().
3. Execute:
   ```shell
   go run main.go
   ```

Лицензия
-------

[![license](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](./LICENSE)
