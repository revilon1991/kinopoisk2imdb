package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/agnivade/levenshtein"
	"github.com/rainycape/unidecode"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	_ "github.com/davecgh/go-spew/spew"
)

type MovieKp struct {
	Title       string
	TitleOrigin string
	Year        string
	SelfRating  string
	IDImdb      string
}

type OrderedFilms struct {
	Keys  []int
	Items map[int]FilmImdb
}

type FilmImdb struct {
	Title       string
	TitleOrigin string
	Year        string
	SelfRating  string
	IDImdb      string
	NameImdb    string
	YearImdb    string
}

type Search struct {
	D []struct {
		I struct {
			Height   int    `json:"height"`
			ImageURL string `json:"imageUrl"`
			Width    int    `json:"width"`
		} `json:"i,omitempty"`
		ID   string `json:"id"`
		L    string `json:"l"`
		Q    string `json:"q"`
		Qid  string `json:"qid"`
		Rank int    `json:"rank"`
		S    string `json:"s"`
		V    []struct {
			I struct {
				Height   int    `json:"height"`
				ImageURL string `json:"imageUrl"`
				Width    int    `json:"width"`
			} `json:"i"`
			ID string `json:"id"`
			L  string `json:"l"`
			S  string `json:"s"`
		} `json:"v,omitempty"`
		Vt int `json:"vt,omitempty"`
		Y  int `json:"y,omitempty"`
	} `json:"d"`
	Q string `json:"q"`
	V int    `json:"v"`
}

const readCsv = "watched_movies.csv"
const writeCsv = "watched_movies.csv"

const KpCookieYaSessId = "0:0000000000.0.0.0000000000000:xx-XXx:000x.0.0:0|000000000.0.0.0:0000000000|000000000.0000000.0.0:0000000.0:0000000000|00:00000000.00000.XXxX0XXXXxXXXXXxxx0xxX0XXXX"
const KpWatchListId = 0
const KpFavoriteListId = 0

const ImdbCookieAtMain = "Xxxx|XxXXXXxXXxXXXxXXXxx0XXxxXX0X0XXXxxX0X0_XxxXXx0X0X0xxxXxXXxxXXXxXXx_xX0x_XXX0xX_x-xXXxxxXxxX0xxxxxX0X0XxX0XxXXxXx0Xxxx00X_XXX-XXx0XxXXXxXxXxxXxXx0xxx00Xx0XxXXxxxXXXXxX0x00XXxx0xXXXxX0xXxX0XXXXxX0XxxXxXX0xXxxXxXx0XXxXx00XXxX0xXXXxx0x00xxXXXxxxx"
const ImdbCookieUbidMain = "000-0000000-0000000"
const ImdbCheckInListId = "ls000000000"
const ImdbFavoriteListId = "ls000000000"

func main() {
	//kpListParser(KpWatchListId)
	//kpListParser(KpFavoriteListId)
	//kpWatchedListParser()

	//imdbMapping()
	//showDuplicates()

	//addFilmToImdbList(ImdbCheckInListId, 0)
	//addFilmToImdbList(ImdbFavoriteListId, 0)
	//addFilmToImdbWatchList(0)

	//rateFilmToImdb(0)
}

func rateFilmToImdb(fromLineExclusive int) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	rateFilm := func(filmId string, selfRating string) {
		u, _ := url.Parse("https://api.graphql.imdb.com/")
		jar.SetCookies(u, []*http.Cookie{
			{Name: "at-main", Value: ImdbCookieAtMain},
			{Name: "ubid-main", Value: ImdbCookieUbidMain},
		})

		dataStr := fmt.Sprintf(
			`{"query":"mutation UpdateTitleRating($rating: Int!, $titleId: ID!) {\n  rateTitle(input: {rating: $rating, titleId: $titleId}) {\n    rating {\n      value\n    }\n  }\n}","operationName":"UpdateTitleRating","variables":{"rating":%s,"titleId":"%s"}}`,
			selfRating,
			filmId,
		)
		var data = strings.NewReader(dataStr)
		req, err := http.NewRequest("POST", "https://api.graphql.imdb.com/", data)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if strings.Contains(string(bodyText), `"errors"`) {
			log.Fatal(string(bodyText))
		}
	}

	movies, err := readCSV(readCsv, 0)
	if err != nil {
		panic(err)
	}

	for _, i := range movies.Keys {
		movie := movies.Items[i]

		rating := strings.TrimSpace(movie.SelfRating)

		if movie.IDImdb == "" || movie.IDImdb == "IDImdb" || !isPositiveNumber(rating) {
			continue
		}

		if i <= fromLineExclusive-1 {
			continue
		}

		rateFilm(movie.IDImdb, rating)

		fmt.Println(fmt.Sprintf(">line %d | %s (%s) rated as %s", i+1, movie.Title, movie.Year, rating))
	}
}

func isPositiveNumber(s string) bool {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return n > 0
}

func addFilmToImdbWatchList(fromLineExclusive int) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	addFilm := func(filmId string) {
		u, _ := url.Parse("https://www.imdb.com/")
		jar.SetCookies(u, []*http.Cookie{
			{Name: "at-main", Value: ImdbCookieAtMain},
			{Name: "ubid-main", Value: ImdbCookieUbidMain},
		})

		finishedUrl := fmt.Sprintf(`https://www.imdb.com/watchlist/%s`, filmId)
		var data = strings.NewReader("")
		req, err := http.NewRequest("PUT", finishedUrl, data)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if !strings.Contains(string(bodyText), `"status":200`) {
			log.Fatal(string(bodyText))
		}
	}

	movies, err := readCSV(readCsv, 0)
	if err != nil {
		panic(err)
	}

	for _, i := range movies.Keys {
		movie := movies.Items[i]
		if movie.IDImdb == "" || movie.IDImdb == "IDImdb" {
			continue
		}

		if i <= fromLineExclusive-1 {
			continue
		}

		addFilm(movie.IDImdb)

		fmt.Println(fmt.Sprintf(">line %d | %s (%s) added", i+1, movie.Title, movie.Year))
	}
}

func addFilmToImdbList(listId string, fromLineExclusive int) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	addFilm := func(filmId string) {
		u, _ := url.Parse("https://api.graphql.imdb.com/")
		jar.SetCookies(u, []*http.Cookie{
			{Name: "at-main", Value: ImdbCookieAtMain},
			{Name: "ubid-main", Value: ImdbCookieUbidMain},
		})

		dataStr := fmt.Sprintf(
			`{"query":"mutation AddConstToList($listId: ID!, $constId: ID!, $includeListItemMetadata: Boolean!, $refTagQueryParam: String, $originalTitleText: Boolean, $isInPace: Boolean! = false) {\n  addItemToList(input: {listId: $listId, item: {itemElementId: $constId}}) {\n    listId\n    modifiedItem {\n      ...EditListItemMetadata\n      listItem @include(if: $includeListItemMetadata) {\n        ... on Title {\n          ...TitleListItemMetadata\n        }\n        ... on Name {\n          ...NameListItemMetadata\n        }\n        ... on Image {\n          ...ImageListItemMetadata\n        }\n        ... on Video {\n          ...VideoListItemMetadata\n        }\n      }\n    }\n  }\n}\n\nfragment EditListItemMetadata on ListItemNode {\n  itemId\n  createdDate\n  absolutePosition\n  description {\n    originalText {\n      markdown\n      plaidHtml(showLineBreak: true)\n      plainText\n    }\n  }\n}\n\nfragment TitleListItemMetadata on Title {\n  ...TitleListItemMetadataEssentials\n  latestTrailer {\n    id\n  }\n  plot {\n    plotText {\n      plainText\n    }\n  }\n  releaseDate {\n    day\n    month\n    year\n  }\n  productionStatus {\n    currentProductionStage {\n      id\n      text\n    }\n  }\n}\n\nfragment TitleListItemMetadataEssentials on Title {\n  ...BaseTitleCard\n  series {\n    series {\n      id\n      originalTitleText {\n        text\n      }\n      releaseYear {\n        endYear\n        year\n      }\n      titleText {\n        text\n      }\n    }\n  }\n}\n\nfragment BaseTitleCard on Title {\n  id\n  titleText {\n    text\n  }\n  titleType {\n    id\n    text\n    canHaveEpisodes\n    displayableProperty {\n      value {\n        plainText\n      }\n    }\n  }\n  originalTitleText {\n    text\n  }\n  primaryImage {\n    id\n    width\n    height\n    url\n    caption {\n      plainText\n    }\n  }\n  releaseYear {\n    year\n    endYear\n  }\n  ratingsSummary {\n    aggregateRating\n    voteCount\n  }\n  runtime {\n    seconds\n  }\n  certificate {\n    rating\n  }\n  canRate {\n    isRatable\n  }\n  titleGenres {\n    genres(limit: 3) {\n      genre {\n        text\n      }\n    }\n  }\n}\n\nfragment NameListItemMetadata on Name {\n  id\n  primaryImage {\n    url\n    caption {\n      plainText\n    }\n    width\n    height\n  }\n  nameText {\n    text\n  }\n  primaryProfessions {\n    category {\n      text\n    }\n  }\n  professions {\n    profession {\n      text\n    }\n  }\n  knownForV2(limit: 1) @include(if: $isInPace) {\n    credits {\n      title {\n        id\n        originalTitleText {\n          text\n        }\n        titleText {\n          text\n        }\n        titleType {\n          canHaveEpisodes\n        }\n        releaseYear {\n          year\n          endYear\n        }\n      }\n      episodeCredits(first: 0) {\n        yearRange {\n          year\n          endYear\n        }\n      }\n    }\n  }\n  knownFor(first: 1) {\n    edges {\n      node {\n        summary {\n          yearRange {\n            year\n            endYear\n          }\n        }\n        title {\n          id\n          originalTitleText {\n            text\n          }\n          titleText {\n            text\n          }\n          titleType {\n            canHaveEpisodes\n          }\n        }\n      }\n    }\n  }\n  bio {\n    displayableArticle {\n      body {\n        plaidHtml(\n          queryParams: $refTagQueryParam\n          showOriginalTitleText: $originalTitleText\n        )\n      }\n    }\n  }\n}\n\nfragment ImageListItemMetadata on Image {\n  id\n  url\n  height\n  width\n  caption {\n    plainText\n  }\n  names(limit: 4) {\n    id\n    nameText {\n      text\n    }\n  }\n  titles(limit: 1) {\n    id\n    titleText {\n      text\n    }\n    originalTitleText {\n      text\n    }\n    releaseYear {\n      year\n      endYear\n    }\n  }\n}\n\nfragment VideoListItemMetadata on Video {\n  id\n  thumbnail {\n    url\n    width\n    height\n  }\n  name {\n    value\n    language\n  }\n  description {\n    value\n    language\n  }\n  runtime {\n    unit\n    value\n  }\n  primaryTitle {\n    id\n    originalTitleText {\n      text\n    }\n    titleText {\n      text\n    }\n    titleType {\n      canHaveEpisodes\n    }\n    releaseYear {\n      year\n      endYear\n    }\n  }\n}","operationName":"AddConstToList","variables":{"listId":"%s","constId":"%s","includeListItemMetadata":false,"refTagQueryParam":"tt_ov_lst","originalTitleText":true,"isInPace":false}}`,
			listId,
			filmId,
		)
		var data = strings.NewReader(dataStr)
		req, err := http.NewRequest("POST", "https://api.graphql.imdb.com/", data)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if strings.Contains(string(bodyText), `"errors"`) {
			log.Fatal(string(bodyText))
		}
	}

	movies, err := readCSV(readCsv, 0)
	if err != nil {
		panic(err)
	}

	for _, i := range movies.Keys {
		movie := movies.Items[i]
		if movie.IDImdb == "" || movie.IDImdb == "IDImdb" {
			continue
		}

		if i <= fromLineExclusive-1 {
			continue
		}

		addFilm(movie.IDImdb)

		fmt.Println(fmt.Sprintf(">line %d | %s (%s) added", i+1, movie.Title, movie.Year))
	}
}

func imdbMapping() {
	movies, err := readCSV(readCsv, 0)
	if err != nil {
		panic(err)
	}

	searchFilm := func(query string) Search {
		jar, _ := cookiejar.New(nil)
		client := &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
		}
		u, _ := url.Parse("https://v3.sg.media-imdb.com/")
		jar.SetCookies(u, []*http.Cookie{})

		escaped := url.QueryEscape(query)
		finalURL := fmt.Sprintf("https://v3.sg.media-imdb.com/suggestion/titles/x/%s.json?includeVideos=0", escaped)

		for {
			resp, err := client.Get(finalURL)
			if err != nil {
				if strings.Contains(err.Error(), "Client.Timeout") || strings.Contains(err.Error(), "timeout") {
					fmt.Println("⚠️  Timeout while request, retry in 1 second...")
					time.Sleep(1 * time.Second)
					continue
				}
				panic(err)
			}

			defer resp.Body.Close()

			var search Search
			err = json.NewDecoder(resp.Body).Decode(&search)
			if err != nil {
				panic(err)
			}

			if len(search.D) > 1 && (search.D[0].ID == "/features/kleenexscore/" || search.D[0].ID == "/imdbpicks/summer-watch-guide/") {
				search.D = search.D[1:]
			}

			return search
		}
	}

	saveToStruct := func(search Search, movie FilmImdb, idx int) FilmImdb {
		movie.IDImdb = search.D[idx].ID
		movie.NameImdb = search.D[idx].L
		movie.YearImdb = strconv.Itoa(search.D[idx].Y)
		return movie
	}

outer:
	for _, i := range movies.Keys {
		movie := movies.Items[i]
		if movie.IDImdb != "" {
			continue
		}

		titleRus := strings.ReplaceAll(movie.Title, "(сериал)", "")
		titleEng := strings.ReplaceAll(movie.TitleOrigin, "(сериал)", "")
		year := ExtractFirst4Digits(movie.Year)

		query := titleEng + " " + strconv.Itoa(year)
		searchEngYear := searchFilm(query)

		// query - eng + year
		// exact year
		// fuzzy en title
		for j := 0; j <= 3; j++ {
			if j >= len(searchEngYear.D) {
				break
			}
			if year == searchEngYear.D[j].Y && FuzzyEqual(searchEngYear.D[j].L, movie.TitleOrigin) {
				movie = saveToStruct(searchEngYear, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		query = titleRus + " " + strconv.Itoa(year)
		searchRusYear := searchFilm(query)

		// query - rus + year
		// exact year
		// fuzzy ru title
		for j := 0; j <= 3; j++ {
			if j >= len(searchRusYear.D) {
				break
			}
			if year == searchRusYear.D[j].Y && FuzzyEqual(searchRusYear.D[j].L, movie.Title) {
				movie = saveToStruct(searchRusYear, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		// query - eng + year
		// aprox year
		// fuzzy en title
		for j := 0; j <= 3; j++ {
			if j >= len(searchEngYear.D) {
				break
			}
			if IsNotTooFarYear(searchEngYear.D[j].Y, year) && FuzzyEqual(searchEngYear.D[j].L, movie.TitleOrigin) {
				movie = saveToStruct(searchEngYear, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		// query - rus + year
		// aprox year
		// fuzzy ru title
		for j := 0; j <= 3; j++ {
			if j >= len(searchRusYear.D) {
				break
			}
			if IsNotTooFarYear(searchRusYear.D[j].Y, year) && FuzzyEqual(searchRusYear.D[j].L, movie.Title) {
				movie = saveToStruct(searchRusYear, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		// query - rus + year
		// aprox year
		// fuzzy ru unicode title
		for j := 0; j <= 3; j++ {
			if j >= len(searchRusYear.D) {
				break
			}
			titleImdbSanitized := OnlyLettersDigitsAndSpaces(searchRusYear.D[j].L)
			titleKpSanitized := OnlyLettersDigitsAndSpaces(movie.Title)
			titleImdbUnidecode := unidecode.Unidecode(titleImdbSanitized)
			titleKpUnidecode := unidecode.Unidecode(titleKpSanitized)
			titleImdbUnidecode = OnlyLettersDigitsAndSpaces(titleImdbUnidecode)

			if IsNotTooFarYear(searchRusYear.D[j].Y, year) && FuzzyEqual(titleImdbUnidecode, titleKpUnidecode) {
				movie = saveToStruct(searchRusYear, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		// query - eng + year
		// aprox year
		// fuzzy en unicode title
		for j := 0; j <= 3; j++ {
			if j >= len(searchEngYear.D) {
				break
			}

			titleImdbSanitized := OnlyLettersDigitsAndSpaces(searchEngYear.D[j].L)
			titleKpSanitized := OnlyLettersDigitsAndSpaces(movie.Title)
			titleImdbUnidecode := unidecode.Unidecode(titleImdbSanitized)
			titleKpUnidecode := unidecode.Unidecode(titleKpSanitized)
			titleImdbUnidecode = OnlyLettersDigitsAndSpaces(titleImdbUnidecode)

			if IsNotTooFarYear(searchEngYear.D[j].Y, year) && FuzzyEqual(titleImdbUnidecode, titleKpUnidecode) {
				movie = saveToStruct(searchEngYear, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		titleKpUnidecode := unidecode.Unidecode(titleRus)
		titleKpUnidecode = OnlyLettersDigitsAndSpaces(titleKpUnidecode)
		query = titleKpUnidecode + " " + strconv.Itoa(year)
		searchRusTranscribedYear := searchFilm(query)

		// query - rus-transcribed + year
		// aprox year
		// fuzzy en unicode title
		for j := 0; j <= 3; j++ {
			if j >= len(searchRusTranscribedYear.D) {
				break
			}

			if IsNotTooFarYear(searchRusTranscribedYear.D[j].Y, year) && FuzzyEqual(searchRusTranscribedYear.D[j].L, titleKpUnidecode) {
				movie = saveToStruct(searchRusTranscribedYear, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		// query - rus + year
		// aprox year
		// fuzzy en title
		for j := 0; j <= 3; j++ {
			if j >= len(searchRusYear.D) {
				break
			}
			if IsNotTooFarYear(searchRusYear.D[j].Y, year) && FuzzyEqual(searchRusYear.D[j].L, movie.TitleOrigin) {
				movie = saveToStruct(searchRusYear, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		// query - eng + year
		// aprox year
		for j := 0; j <= 3; j++ {
			if j >= len(searchEngYear.D) {
				break
			}
			if IsNotTooFarYear(searchEngYear.D[j].Y, year) {
				movie = saveToStruct(searchEngYear, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		// query - rus + year
		// aprox year
		for j := 0; j <= 3; j++ {
			if j >= len(searchRusYear.D) {
				break
			}
			if IsNotTooFarYear(searchRusYear.D[j].Y, year) {
				movie = saveToStruct(searchRusYear, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		query = titleEng
		searchEng := searchFilm(query)

		// query - eng
		// aprox year
		for j := 0; j <= 3; j++ {
			if j >= len(searchEng.D) {
				break
			}
			if IsNotTooFarYear(searchEng.D[j].Y, year) {
				movie = saveToStruct(searchEng, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		query = titleRus
		searchRus := searchFilm(query)

		// query - rus
		// aprox year
		for j := 0; j <= 3; j++ {
			if j >= len(searchRus.D) {
				break
			}
			if IsNotTooFarYear(searchRus.D[j].Y, year) {
				movie = saveToStruct(searchRus, movie, j)
				movies.Items[i] = movie
				fmt.Println(fmt.Sprintf(">%d line | %s (%s) mapped", i, movie.Title, movie.Year))
				continue outer
			}
		}

		fmt.Println(fmt.Sprintf("No results found for %s (%s)", movie.TitleOrigin, movie.Year))
	}

	_ = writeCSV(writeCsv, movies)
}

func OnlyLettersDigitsAndSpaces(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func ExtractFirst4Digits(s string) int {
	re := regexp.MustCompile(`\d{4}`)
	match := re.FindString(s)

	value, _ := strconv.Atoi(match)

	return value
}

func IsNotTooFarYear(year1, year2 int) bool {
	diff := year1 - year2
	if diff < 0 {
		diff = -diff
	}

	return diff <= 1
}

func showDuplicates() {
	file, err := os.Open(readCsv)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true

	idOccurrences := make(map[string][]int)

	lineNumber := 1
	for {
		record, err := reader.Read()
		if err != nil {
			break // либо обработать EOF/ошибку
		}

		if len(record) > 4 {
			id := record[4]
			if id != "" {
				idOccurrences[id] = append(idOccurrences[id], lineNumber)
			}
		}

		lineNumber++
	}

	for id, lines := range idOccurrences {
		if len(lines) > 1 {
			fmt.Printf("ID %s found on rows: %v\n", id, lines)
		}
	}
}

func normalizeAndSort(s string) string {
	re := regexp.MustCompile(`[^\p{L}\p{N}\s]`)
	s = re.ReplaceAllString(s, "")
	s = strings.ToLower(s)

	words := strings.Fields(s)
	sort.Strings(words)

	return strings.Join(words, " ")
}

func FuzzyEqual(a, b string) bool {
	na := normalizeAndSort(a)
	nb := normalizeAndSort(b)

	if na == nb {
		return true
	}

	dist := levenshtein.ComputeDistance(na, nb)
	maxAllowed := len(na) / 4

	return dist <= maxAllowed
}

func readCSV(path string, count int) (OrderedFilms, error) {
	file, err := os.Open(path)
	if err != nil {
		return OrderedFilms{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return OrderedFilms{}, err
	}

	items := map[int]FilmImdb{}
	var keys []int

	cnt := 0
	for i, row := range records {
		if i == 0 {
			continue
		}
		if count != 0 && cnt >= count {
			break
		}
		cnt++

		f := FilmImdb{
			Title:       row[0],
			TitleOrigin: row[1],
			Year:        row[2],
			SelfRating:  row[3],
		}

		if len(row) > 4 {
			f.IDImdb = row[4]
			f.NameImdb = row[5]
			f.YearImdb = row[6]
		}

		items[i] = f
		keys = append(keys, i)
	}

	return OrderedFilms{Keys: keys, Items: items}, nil
}

func writeCSV(path string, films OrderedFilms) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	_ = writer.Write([]string{"Title", "TitleOrigin", "Year", "SelfRating", "IDImdb", "NameImdb", "YearImdb"})

	for _, i := range films.Keys {
		f := films.Items[i]
		_ = writer.Write([]string{f.Title, f.TitleOrigin, f.Year, f.SelfRating, f.IDImdb, f.NameImdb, f.YearImdb})
	}
	return nil
}

func findImdbID(films []FilmImdb, titleEng string) string {
	for _, f := range films {
		if f.TitleOrigin == titleEng {
			return f.IDImdb
		}
	}
	return ""
}

func kpWatchedListParser() {
	var movies = map[string]MovieKp{}

	for page := 1; ; page++ {
		if _, err := os.Stat(fmt.Sprintf("pages/%d.html", page)); os.IsNotExist(err) {
			break
		}

		file, err := os.Open(fmt.Sprintf("pages/%d.html", page))
		if err != nil {
			fmt.Println("Error while opening file:", err)
			return
		}
		defer file.Close()

		doc, _ := goquery.NewDocumentFromReader(file)

		selection := doc.Find(".profileFilmsList .item").Each(func(i int, s *goquery.Selection) {
			titleRusMeta := s.Find(".info .nameRus a").Text()
			titleEng := s.Find(".info .nameEng").Text()
			titleRus, year := parseTitleAndYear(titleRusMeta)
			if isEmptyOrNbspOnly(titleEng) {
				titleEng = titleRus
			}

			filmId := extractFilmId(s.Text())

			titleRus = replaceNbsp(titleRus)
			titleEng = replaceNbsp(titleEng)

			if movie, ok := movies[filmId]; ok {
				movies[filmId] = movie
			} else {
				movies[filmId] = MovieKp{Title: titleRus, TitleOrigin: titleEng, Year: year}
			}
		})

		if selection.Length() == 0 {
			fmt.Println("Element did not founded, break loop.")
			break
		}

		fmt.Println(fmt.Sprintf("Page %d done", page))
	}

	_ = saveKpToCSV(movies, writeCsv)
}

func replaceNbsp(s string) string {
	return strings.ReplaceAll(s, "\u00A0", " ")
}

// parse only favorite or will watch films
func kpListParser(folder int) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Jar: jar,
	}

	u, _ := url.Parse("https://www.kinopoisk.ru")

	jar.SetCookies(u, []*http.Cookie{
		{Name: "ya_sess_id", Value: KpCookieYaSessId},
	})

	filmCounter := 0

	var movies = map[string]MovieKp{}

	for page := 1; ; page++ {
		fmt.Println(fmt.Sprintf("Start request for page %d", page))

		resp, err := client.Get(fmt.Sprintf("https://www.kinopoisk.ru/mykp/folders/%d/?page=%d&limit=200", folder, page))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		fmt.Println(fmt.Sprintf("Request for page %d done", page))

		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		selection := doc.Find(".item").Each(func(i int, s *goquery.Selection) {
			titleRus := s.Find(".info a.name").Text()
			titleEngMeta := s.Find(".info span:nth-of-type(1)").Text()
			titleEng, year := parseTitleAndYear(titleEngMeta)
			if titleEng == "" {
				titleEng = titleRus
			}
			rating, _ := extractUserRating(s.Text())

			filmCounter++
			filmCounterStr := strconv.Itoa(filmCounter)
			movies[filmCounterStr] = MovieKp{Title: titleRus, TitleOrigin: titleEng, Year: year, SelfRating: rating}
		})

		if selection.Length() == 0 {
			fmt.Println("Element did not founded, break loop.")
			break
		}

		err = saveKpToCSV(movies, writeCsv)

		fmt.Println(fmt.Sprintf("Page %d done", page))
	}
}

func isEmptyOrNbspOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) && r != '\u00A0' {
			return false
		}
	}
	return true
}

func extractFilmId(html string) string {
	const key = "film: "
	start := strings.Index(html, key)
	if start == -1 {
		panic("cant find film id:" + html)
	}
	start += len(key)
	end := strings.Index(html[start:], ",")
	if end == -1 {
		panic("cant find film id:" + html)
	}
	return html[start : start+end]
}

func extractUserRating(html string) (string, bool) {
	const key = "rating: '"
	start := strings.Index(html, key)
	if start == -1 {
		return "", false
	}
	start += len(key)
	end := strings.Index(html[start:], "'")
	if end == -1 {
		return "", false
	}
	return html[start : start+end], true
}

func parseTitleAndYear(s string) (title, year string) {
	start := strings.Index(s, "(")
	end := strings.Index(s, ")")

	if start == -1 || end == -1 || end <= start {
		title = strings.TrimSpace(s)
		return
	}

	title = strings.TrimSpace(s[:start])
	year = strings.TrimSpace(s[start+1 : end])

	return
}

func saveKpToCSV(movies map[string]MovieKp, fileName string) error {
	_, err := os.Stat(fileName)
	fileExists := err == nil

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !fileExists {
		err = writer.Write([]string{"Title", "TitleOrigin", "Year", "SelfRating", "IDImdb"})
		if err != nil {
			return err
		}
	}

	for _, movie := range movies {
		err := writer.Write([]string{movie.Title, movie.TitleOrigin, movie.Year, movie.SelfRating, movie.IDImdb})
		if err != nil {
			return err
		}
	}

	return nil
}
