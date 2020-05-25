package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

func main() {

	run()
}

func run() {
	r := mux.NewRouter()

	// rest
	r.HandleFunc("/", login)
	r.HandleFunc("/repos", repos)
	r.HandleFunc("/createWebHook", createWebHook)
	r.HandleFunc("/webhook", webhook)
	r.HandleFunc("/logs", showLogs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, r)
}

func login(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")

	body := struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
	}{
		ClientId:     "5c2def9cf8914f910eb8",
		ClientSecret: "b05651cd6cd3886957231dbd68e68c6629aab19c",
		Code:         code,
	}

	data, err := json.Marshal(&body)
	if err != nil {
		writeError(w, err)
		return
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(data))
	if err != nil {
		writeError(w, err)
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		writeError(w, err)
		return
	}

	type responseStruct struct {
		AccessToken string `json:"access_token"`
	}

	var response responseStruct

	responseBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		writeError(w, err)
		return
	}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		writeError(w, err)
		return
	}

	result := struct {
		AccessToken string
	}{
		AccessToken: response.AccessToken,
	}

	write(w, result)
}

func repos(w http.ResponseWriter, r *http.Request) {

	accessToken := r.URL.Query().Get("accessToken")
	if accessToken == "" {
		writeError(w, errors.New("accessToken missing"))
		return
	}

	user := r.URL.Query().Get("user")
	if user == "" {
		writeError(w, errors.New("user missing"))
		return
	}

	ghClient := createGHClient(accessToken)

	repos, _, err := ghClient.Repositories.List(context.Background(), user, nil)
	if err != nil {
		writeError(w, err)
		return
	}

	result := struct {
		Repos []string
	}{
		Repos: func(repos []*github.Repository) (res []string) {
			for _, r := range repos {
				res = append(res, *r.Name)
			}
			return
		}(repos),
	}

	write(w, result)
}

func createWebHook(w http.ResponseWriter, r *http.Request) {

	accessToken := r.URL.Query().Get("accessToken")
	if accessToken == "" {
		writeError(w, errors.New("accessToken missing"))
		return
	}

	user := r.URL.Query().Get("user")
	if user == "" {
		writeError(w, errors.New("user missing"))
		return
	}

	repo := r.URL.Query().Get("repo")
	if repo == "" {
		writeError(w, errors.New("repo missing"))
		return
	}

	ghClient := createGHClient(accessToken)

	active := true

	config := make(map[string]interface{})
	config["url"] = "https://githubtest12345.herokuapp.com/webhook"
	config["content_type"] = "json"
	config["insecure_ssl"] = "0"

	_, _, err := ghClient.Repositories.CreateHook(context.Background(), user, repo, &github.Hook{
		Config: config,
		Events: []string{
			"push",
		},
		Active: &active,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeOk(w)
}

func webhook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, err)
		return
	}

	f, err := os.OpenFile("./log.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		writeError(w, err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(time.Now().Format(time.RFC3339) + ": " + string(body) + "\n\n"); err != nil {
		writeError(w, err)
		return
	}

	result := struct {
		Body string
	}{
		Body: string(body),
	}

	write(w, result)
}

func showLogs(w http.ResponseWriter, r *http.Request) {

	data, err := ioutil.ReadFile("./log.txt")
	if err != nil {
		writeError(w, err)
		return
	}

	result := struct {
		Logs string
	}{
		Logs: string(data),
	}

	write(w, result)
}

func write(w http.ResponseWriter, object interface{}) {
	data, err := json.Marshal(object)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)

	w.Write([]byte(err.Error()))

}

func writeOk(w http.ResponseWriter) {
	result := struct {
		Success bool
	}{
		Success: true,
	}

	write(w, result)
}

func createGHClient(accessToken string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}
