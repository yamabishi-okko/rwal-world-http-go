package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// レスポンス用モデル（PAY.JPの代わりにjsonplaceholderを使う）
type Post struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID int    `json:"userId"`
}

// 簡単なAPIクライアント
type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTP:    &http.Client{Timeout: 5 * time.Second},
	}
}

// 認証つきGETリクエスト（PAY.JPはAuthorizationヘッダが必須）
func (c *Client) doGet(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	// PAY.JPなら Bearer sk_test_xxx
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	return c.HTTP.Do(req)
}

func main() {
	// ダミーAPIを利用（PAY.JPの代わり）
	c := NewClient("https://jsonplaceholder.typicode.com", "sk_test_dummy")

	ctx := context.Background()

	// 顧客一覧を取る → 今回は /posts に置き換え
	// クライアントがGETリクエストを送っている
	resp, err := c.doGet(ctx, "/posts?userId=1")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		panic(fmt.Errorf("status=%d body=%s", resp.StatusCode, string(b)))
	}

	var posts []Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		panic(err)
	}

	fmt.Printf("== 顧客一覧（ダミーAPI） ==\n")
	for _, p := range posts {
		fmt.Printf("ID=%d Title=%s\n", p.ID, p.Title)
	}
}
