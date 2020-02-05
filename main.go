package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

var (
	list = []string{}
)

func main() {
	f, err := os.Open("./assets/words.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}

	r := chi.NewRouter()
	r.Get("/", mainHandler)
	r.Get("/words", wordsHandler)

	fileServer := http.StripPrefix("/static", http.FileServer(http.Dir("./static")))
	r.Mount("/static", fileServer)

	http.ListenAndServe(":7070", r)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	f, _ := os.Open("./static/index.html")
	defer f.Close()

	d, _ := ioutil.ReadAll(f)

	w.Write(d)
	//http.Redirect(w, r, "https://www.discord.gg/chiaki", http.StatusTemporaryRedirect)
}

func wordsHandler(w http.ResponseWriter, r *http.Request) {

	inp := r.FormValue("query")
	limit := r.FormValue("limit")
	if limit == "" {
		limit = "50"
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 50
	}

	if limitInt > 1000 {
		limitInt = 1000
	}

	t1 := time.Now()
	respList := findMatches(inp, list)
	total := time.Now().Sub(t1)
	fmt.Println("word: ", inp, "\ttime taken: ", total.String(), "\tresults found: ", len(respList))

	sort.Sort(SortByCloseness(respList))

	g := make([]string, int64(math.Min(float64(len(respList)), float64(limitInt))))
	for i := range g {
		g[i] = respList[i].Word
	}

	d, _ := json.Marshal(g)

	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.Write(d)
}

func findMatches(input string, list []string) []Word {
	if input == "" {
		return []Word{}
	}

	input = strings.ToLower(input)

	out := []Word{}

	for _, word := range list {

		// if input is longer then go next
		if len(input) > len(word) {
			continue
		}

		w := strings.ToLower(word)

		// total letters in word
		total := len(w)

		// last index of the current word
		lastIndex := 0

		// total characters from input matching the word
		totalFound := 0

		// check at least the first letter of the input word to see if it matches any letters in the word, if it matches,
		// then keep going, if there is no matches found, then the word doesnt match
		for i := 0; i < len(input); i++ {
			inpChar := input[i]

			for j := lastIndex; j < total; j++ {
				if w[j] == inpChar {
					lastIndex = j + 1
					totalFound++
					break
				}
			}

			// if first letter of input wasnt found in the whole word, this aint it chief
			if totalFound == 0 {
				break
			}
		}

		// if all letters of the input was found in the word, add to output list
		if totalFound == len(input) {
			out = append(out, Word{
				Word:      word,
				Closeness: len(word) - len(input),
			})
		}
	}
	return out
}

type SortByCloseness []Word

func (a SortByCloseness) Len() int           { return len(a) }
func (a SortByCloseness) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByCloseness) Less(i, j int) bool { return a[i].Closeness < a[j].Closeness }

type Word struct {
	Word      string `json:"word"`
	Closeness int    `json:"-"`
}
