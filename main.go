package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"./env"
	"github.com/ChimeraCoder/anaconda"
	"github.com/gorilla/mux"
	"github.com/mrjones/oauth"
	"github.com/pocke/hlog.go"
	"github.com/yosssi/ace"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/get_token", twitterGetTokenHandler)
	r.HandleFunc("/callback", twitterCallbackHandler)
	r.HandleFunc("/css/bootstrap.min.css", handleAssets("assets/bootstrap.min.css"))

	http.HandleFunc("/", hlog.Wrap(r.ServeHTTP))

	log.Println("start web server")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("root")

	tpl, err := ace.Load("assets/main", "", &ace.Options{
		DynamicReload: env.DEBUG,
		Asset:         Asset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleAssets(path string) http.HandlerFunc {
	var t string
	if strings.HasSuffix(path, "css") {
		t = "text/css"
	} else if strings.HasSuffix(path, "js") {
		t = "application/javascript"
	} else if strings.HasSuffix(path, "html") {
		t = "text/html"
	} else {
		t = "text/plane"
	}

	var data []byte
	var err error
	if !env.DEBUG {
		data, err = Asset(path)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if env.DEBUG {
			data, err = Asset(path)
		}

		data, err := Asset(path)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", t)
		w.Write(data)
	}
}

func twitterGetTokenHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("get token")

	tokenUrl := fmt.Sprintf("http://%s/callback", r.Host)
	token, reqUrl, err := twitter.GetRequestTokenAndUrl(tokenUrl)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokens[token.Token] = token

	http.Redirect(w, r, reqUrl, http.StatusTemporaryRedirect)
}

func twitterCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("callback")

	values := r.URL.Query()
	verificationCode := values.Get("oauth_verifier")
	tokenKey := values.Get("oauth_token")

	accessToken, err := twitter.AuthorizeToken(tokens[tokenKey], verificationCode)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	api := anaconda.NewTwitterApi(accessToken.Token, accessToken.Secret)
	_, err = api.PostTweet("にゃーん", url.Values{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tpl, err := ace.Load("assets/tweeted", "", &ace.Options{
		DynamicReload: env.DEBUG,
		Asset:         Asset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var tokens = make(map[string]*oauth.RequestToken)
var ck = "Sl2mPb7KTyjEqORYO3VQ"
var cs = "rpWErProPgO0WzPkruXZEwrtfzEJ3EXMJXiPRpT1c"

var twitter = func() *oauth.Consumer {

	return oauth.NewConsumer(
		ck,
		cs,
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		},
	)
}()

func init() {
	anaconda.SetConsumerKey(ck)
	anaconda.SetConsumerSecret(cs)
}
