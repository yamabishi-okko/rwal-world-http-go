package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type User struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

type Email struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type Repo struct {
	Name string `json:"name"`
	Priv bool   `json:"private"`
}

type Client struct {
	base string
	hc   *http.Client
	tok  string
}

func NewClient(token string) *Client {
	return &Client{
		base: "https://api.github.com",
		hc:   &http.Client{Timeout: 10 * time.Second},
		tok:  token,
	}
}

func (c *Client) do(ctx context.Context, method, path string, q url.Values, body io.Reader, out any) (*http.Response, error) {
	u := c.base + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}

	// GitHub REST v3 推奨ヘッダ
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// 認証（PAT）
	if strings.TrimSpace(c.tok) != "" {
		// "token <PAT>" が一般的。Bearer でも可だが token を使う。
		req.Header.Set("Authorization", "token "+c.tok)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return resp, fmt.Errorf("non-2xx %d: %s", resp.StatusCode, string(b))
	}
	if out != nil {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func mustToken() string {
	t := os.Getenv("GITHUB_TOKEN")
	if t == "" {
		panic(errors.New("GITHUB_TOKEN が未設定です。export GITHUB_TOKEN=xxxxx を実行してください"))
	}
	return t
}

func main() {
	ctx := context.Background()
	c := NewClient(mustToken())

	// 1) 自分のユーザー情報（GET /user）
	var me User
	_, err := c.do(ctx, http.MethodGet, "/user", nil, nil, &me)
	must(err)
	fmt.Printf("== Authenticated as: %s (name=%s)\n", me.Login, me.Name)

	// 2) 自分のメール一覧（GET /user/emails）
	var emails []Email
	_, err = c.do(ctx, http.MethodGet, "/user/emails", nil, nil, &emails)
	must(err)
	fmt.Println("== Emails:")
	for _, e := range emails {
		fmt.Printf("- %s primary=%v verified=%v\n", e.Email, e.Primary, e.Verified)
	}

	// 3) 自分のリポジトリを少しだけ（GET /user/repos?per_page=5&sort=updated）
	q := url.Values{
		"per_page": {"5"},
		"sort":     {"updated"},
	}
	var repos []Repo
	resp, err := c.do(ctx, http.MethodGet, "/user/repos", q, nil, &repos)
	must(err)
	fmt.Println("== Repos (latest 5):")
	for _, r := range repos {
		fmt.Printf("- %s (private=%v)\n", r.Name, r.Priv)
	}

	// 4) レート制限の観察（レスポンスヘッダ）
	fmt.Println("== Rate limit headers:")
	fmt.Println("X-RateLimit-Limit:", resp.Header.Get("X-RateLimit-Limit"))
	fmt.Println("X-RateLimit-Remaining:", resp.Header.Get("X-RateLimit-Remaining"))
	fmt.Println("X-RateLimit-Reset(unix):", resp.Header.Get("X-RateLimit-Reset"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
