package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Posts struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type User struct {
	Username  string `json:"username"`
	Passwoord string `json:"password"`
}

var secretKey = []byte("my-secret-key")

func checkAuth(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Connection", "close")
		defer r.Body.Close()

		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return secretKey, nil
			})

			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				w.Header().Add("Content-Type", "application/json")
				return
			}

			if token.Valid {
				endpoint(w, r)
			}

		} else {
			fmt.Fprintf(w, "Not Authorized")
		}
	})
}

func main() {
	r := mux.NewRouter()

	// GET /posts
	r.Handle("/posts", checkAuth(getPosts)).Methods("GET")

	// POST /login
	r.HandleFunc("/login", login).Methods("POST")

	log.Fatal(http.ListenAndServe(":9090", r))
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	var user User
	json.NewDecoder(r.Body).Decode(&user)

	if !isValidUser(user) {
		fmt.Println("Not valid.")
		return
	}

	token, err := GenerateToken()

	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(token)
}

func GenerateToken() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		log.Fatal(err)
	}

	return tokenString, nil
}

func isValidUser(currentUser User) bool {
	if currentUser.Username == "admin" && currentUser.Passwoord == "admin" {
		return true
	}

	return false
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	posts := queryPosts()
	json.NewEncoder(w).Encode(posts)
}

func queryPosts() []Posts {
	var selectedPosts []Posts

	db, err := sql.Open("postgres", "postgres://mrabets:9110@localhost/blogdb?sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("select * from posts")

	for rows.Next() {
		post := Posts{}
		err = rows.Scan(&post.Id, &post.Title, &post.Content)

		if err != nil {
			log.Fatal(err)
		}

		selectedPosts = append(selectedPosts, post)
	}

	return selectedPosts
}
