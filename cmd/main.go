package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	BaseUrl = "https://graph.microsoft.com/beta/me/todo/lists"
)

func main() {
	var err error
	ctx := context.Background()

	conf := &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			AuthStyle: 0,
		},
		RedirectURL: "https://localhost/login/authorized",
		Scopes: []string{
			"Tasks.ReadWrite",
		},
	}

	var token *oauth2.Token
	b, err := os.ReadFile("token.json")
	if err == nil {
		var t oauth2.Token
		err = json.Unmarshal(b, &t)
		if err != nil {
			panic(err)
		}

		if time.Now().Before(t.Expiry) {
			token = &t
		}
	}

	if token == nil {
		token, err = tokenGen(ctx, conf)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("%+v\n", *token)

	defer func() {
		fmt.Printf("%+v\n", *token)
		b, _ = json.Marshal(token)
		err = os.WriteFile("token.json", b, 0o664)
		if err != nil {
			panic(err)
		}
	}()

	client := conf.Client(ctx, token)

	dump(client, "")

}

func dump(client *http.Client, path string) {
	resp, err := client.Get(fmt.Sprintf("%s%s", BaseUrl, path))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

func tokenGen(ctx context.Context, conf *oauth2.Config) (*oauth2.Token, error) {

	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Println("Visit the URL for the auth dialog:", url)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}
