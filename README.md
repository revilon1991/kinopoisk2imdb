# CLI Tool for Transferring Movies from Kinopoisk to IMDb

[![На Русском](https://img.shields.io/badge/Перейти_на-Русский-green.svg?style=flat-square)](./README.RU.md)

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

Each function should be triggered manually by uncommenting it in `main()`.

### 1. Collecting Movies from Kinopoisk

#### `kpListParser()`

Automatically fetches movie data (title, year, rating) from your Kinopoisk lists ("Find on Internet" and "Favorites"). Requires valid `ya-sess-id` cookie.

#### `kpWatchedListParser()`

Requires manual downloading of HTML files:

1. Go to your "Watched Movies" list on Kinopoisk.
2. Save each page as an HTML file:
   - Right-click → Save As → "Webpage, HTML only"
   - Name files `1.html`, `2.html`, etc.
3. Put all files in a folder named `pages` next to `main.go`.
4. Call `kpWatchedListParser()` to extract movie data.

### 2. Mapping to IMDb

#### `imdbMapping()`

- Reads movie data from `readCsv`
- Searches for matching IMDb entries by title/year
- Writes to `writeCsv`, adding:
   - IMDb ID
   - IMDb title
   - IMDb year

> ⚠️ About 5% of mappings may be incorrect or missing. Manual verification is recommended.

#### `showDuplicates()`

Scans `readCsv` for duplicate IMDb IDs to help catch mapping errors.

### 3. Adding Movies to IMDb

#### `addFilmToImdbList(listId string, startRow int)`

Adds movies from `readCsv` to a specified IMDb list. Skips movies with empty IMDb ID. Use `startRow` to resume from a specific row if interrupted.

#### `addFilmToImdbWatchList(startRow int)`

Same as above but adds to the IMDb Watchlist.

#### `rateFilmToImdb()`

Applies user ratings to movies on IMDb if available in `readCsv` (from Kinopoisk).

## ⚠️ Notes

- Mapping accuracy is around 95%.
- Manual review of the `writeCsv` file is **highly recommended**.
- Make sure your cookies and list IDs are valid and up to date.

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

## 📝 License

[![license](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](./LICENSE)
