package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// モデル（JSONPlaceholderの /posts を利用）
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// APIクライアント本体
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string // Bearerトークン（必要なAPIで使う想定。サンプルでは未使用）
}

// コンストラクタ
func NewClient(baseURL string, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // 必須: ハング対策
		},
		token: token,
	}
}

// 汎用: JSONリクエスト送信→JSONレスポンスをデコード
func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	// URL組み立て
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(b))
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// 認証が必要なAPIなら有効化
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// 2xxのみ成功扱い
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("non-2xx status: %d body=%s", resp.StatusCode, string(b))
	}

	if out == nil {
		// ボディ不要の時
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// 一覧取得: GET /posts?userId=1 のようなクエリも体験
func (c *Client) ListPosts(ctx context.Context, userID *int) ([]Post, error) {
	q := url.Values{}
	if userID != nil {
		q.Set("userId", fmt.Sprint(*userID))
	}
	var posts []Post
	if err := c.doJSON(ctx, http.MethodGet, "/posts", q, nil, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// 1件取得: GET /posts/{id}
func (c *Client) GetPost(ctx context.Context, id int) (*Post, error) {
	var p Post
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/posts/%d", id), nil, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// 作成: POST /posts
func (c *Client) CreatePost(ctx context.Context, p *Post) (*Post, error) {
	if p == nil {
		return nil, errors.New("post is nil")
	}
	var created Post
	if err := c.doJSON(ctx, http.MethodPost, "/posts", nil, p, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// 全体更新: PUT /posts/{id}
func (c *Client) UpdatePost(ctx context.Context, id int, p *Post) (*Post, error) {
	if p == nil {
		return nil, errors.New("post is nil")
	}
	var updated Post
	if err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/posts/%d", id), nil, p, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

// 削除: DELETE /posts/{id}
func (c *Client) DeletePost(ctx context.Context, id int) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/posts/%d", id), nil, nil, nil)
}

func main() {
	// ダミーAPI
	c := NewClient("https://jsonplaceholder.typicode.com", "")

	// 各操作はコンテキストでタイムアウトを個別制御可能
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// 1) 一覧取得（クエリ付き）
	fmt.Println("== List posts for userId=1")
	uid := 1
	posts, err := c.ListPosts(ctx, &uid)
	must(err)
	fmt.Printf("got %d posts\n", len(posts))
	if len(posts) > 0 {
		fmt.Printf("first: ID=%d Title=%q\n", posts[0].ID, posts[0].Title)
	}

	// 2) 単一取得
	fmt.Println("\n== Get post 1")
	p1, err := c.GetPost(ctx, 1)
	must(err)
	fmt.Printf("ID=%d Title=%q\n", p1.ID, p1.Title)

	// 3) 作成（POST）
	fmt.Println("\n== Create post")
	newOne := &Post{UserID: 99, Title: "Hello REST", Body: "Created from Go client"}
	created, err := c.CreatePost(ctx, newOne)
	must(err)
	fmt.Printf("created: ID=%d, Title=%q\n", created.ID, created.Title)

	// 4) 更新（PUT）
	fmt.Println("\n== Update post")
	created.Title = "Hello REST (updated)"
	updated, err := c.UpdatePost(ctx, created.ID, created)
	must(err)
	fmt.Printf("updated: ID=%d, Title=%q\n", updated.ID, updated.Title)

	// 5) 削除（DELETE）
	fmt.Println("\n== Delete post")
	if err := c.DeletePost(ctx, created.ID); err != nil {
		must(err)
	}
	fmt.Println("deleted ok")

	// 6) エラー時の挙動確認（存在しないID）
	fmt.Println("\n== Get post 99999 (expect 404)")
	_, err = c.GetPost(ctx, 99999)
	if err != nil {
		fmt.Println("error handled:", err)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
