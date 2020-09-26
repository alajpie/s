package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/julienschmidt/httprouter"
)

func panil(err error) {
	if err != nil {
		panic(err)
	}
}

func handle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conn, err := sqlite3.Open("s.db")
	panil(err)
	defer conn.Close()
	conn.BusyTimeout(5 * time.Second)

	x := ps.ByName("x")[1:]

	if match, _ := regexp.Match(`^https?://`, []byte(x)); match { // create
		err := conn.Exec("INSERT INTO links (link) VALUES (?)", x)
		panil(err)
		id := conn.LastInsertRowID()
		fmt.Fprint(w, "<style>html{font-family:sans-serif;font-size:500%;margin-top:50px;text-align:center}</style>")
		fmt.Fprint(w, id)
		fmt.Fprint(w, "<script>history.pushState({},'','/')</script>") // protect from refreshes wasting ids
	} else { // redirect
		stmt, err := conn.Prepare(`SELECT link FROM links WHERE id = ?`, x)
		panil(err)
		defer stmt.Close()

		hasRow, err := stmt.Step()
		panil(err)
		if !hasRow {
			fmt.Fprint(w, "<style>html{font-family:sans-serif;font-size:500%;margin-top:50px;text-align:center}</style>")
			fmt.Fprint(w, "what")
			return
		}

		var link string
		err = stmt.Scan(&link)
		panil(err)
		http.Redirect(w, r, link, http.StatusMovedPermanently)
	}
}

func main() {
	conn, err := sqlite3.Open("s.db")
	panil(err)
	conn.BusyTimeout(5 * time.Second)
	conn.Exec(`CREATE TABLE IF NOT EXISTS links (id INTEGER PRIMARY KEY, link TEXT)`)
	conn.Close()

	router := httprouter.New()
	router.GET("/*x", handle)

	log.Fatal(http.ListenAndServeTLS(":443", "cert.crt", "cert.key", router))
}
