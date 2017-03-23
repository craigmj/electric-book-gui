package www

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang/glog"

	"ebw/config"
	"ebw/git"
)

type githubToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// githubLogin forwards the user to login on Github
func githubLogin(c *Context) error {
	p := url.Values{}
	p.Add(`client_id`, config.Config.Github.Client)
	p.Add(`redirect_uri`, config.Config.Server+"/github/auth")
	p.Add(`scope`, `repo`)
	p.Add(`state`, `unguessable random string`)
	return c.Render("landing.html", map[string]interface{}{
		"AuthURL": `https://github.com/login/oauth/authorize?` + p.Encode()})
}

// githubAuth receives github's oauth2 authorization response
func githubAuth(w http.ResponseWriter, r *http.Request) {
	p := url.Values{}
	if "" != r.FormValue("error") {
		fmt.Fprintf(w, r.FormValue("error_description"))
		return
	}
	glog.Infof("Callback from github with auth code: %s", r.FormValue("code"))
	p.Add(`code`, r.FormValue(`code`))
	p.Add(`client_id`, config.Config.Github.Client)
	p.Add(`client_secret`, config.Config.Github.Secret)
	p.Add(`redirect_uri`, config.Config.Server+`/github/auth`)
	p.Add(`state`, `unguessable random string`)

	req, err := http.NewRequest("POST", `https://github.com/login/oauth/access_token`,
		strings.NewReader(p.Encode()))
	req.Header.Add("Accept", "application/json")
	if nil != err {
		WebError(w, r, err)
		return
	}
	var client http.Client
	res, err := client.Do(req)
	if nil != err {
		WebError(w, r, err)
		return
	}
	defer res.Body.Close()
	token := &githubToken{}
	if err := json.NewDecoder(res.Body).Decode(&token); nil != err {
		WebError(w, r, err)
	}
	glog.Infof("Setting cookie %s to %s", git.GithubTokenCookie, token.AccessToken)

	http.SetCookie(w, &http.Cookie{
		Name:  git.GithubTokenCookie,
		Value: token.AccessToken,
		Path:  "/",
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func LogoffHandler(c *Context) error {
	client, err := git.ClientFromWebRequest(c.W, c.R)
	if nil != err {
		return err
	}
	if nil == client {
		return nil
	}

	http.SetCookie(c.W, &http.Cookie{
		Name:    git.GithubTokenCookie,
		Value:   ``,
		Path:    `/`,
		Expires: time.Time{},
	})

	c.Redirect(`/to-github?u=%s`, url.QueryEscape(client.Username))
	return nil
}

func ToGithubHandler(c *Context) error {
	return c.Redirect(`https://github.com/%s`, c.P(`u`))
}

func Client(w http.ResponseWriter, r *http.Request) *git.Client {
	client, err := git.ClientFromWebRequest(w, r)
	if err == git.ErrNotLoggedIn {
		http.Redirect(w, r, "/github/login", http.StatusFound)
		return nil
	}
	if nil != err {
		WebError(w, r, err)
		return nil
	}
	return client
}
