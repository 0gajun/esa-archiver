package esa

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/pkg/errors"
)

const (
	apiEndPoint = "https://api.esa.io/v1"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client

	accessToken string

	teamName string
}

func NewEsa(accessToken, teamName string) (*Client, error) {
	if len(accessToken) == 0 {
		return nil, errors.New("missing access token")
	}

	if len(teamName) == 0 {
		return nil, errors.New("missing team name")
	}

	esaUrl, err := url.ParseRequestURI(apiEndPoint)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("esa's api endpoint url is invalid format. (%s)", apiEndPoint))
	}

	return &Client{
		baseURL:     esaUrl,
		httpClient:  &http.Client{},
		accessToken: accessToken,
		teamName:    teamName,
	}, nil
}

type Posts struct {
	Posts []Post `json:"posts"`

	PrevPage   *int `json:"prev_page"`
	NextPage   *int `json:"next_page"`
	TotalCount int  `json:"total_count"`
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	MaxPerPage int  `json:"max_per_page"`
}

const defaultPostSliceCapacity = 1000

func (c *Client) GetAllPosts(ctx context.Context) ([]Post, error) {
	var page = 1
	allPosts := make([]Post, defaultPostSliceCapacity)

	for {
		posts, err := c.getPosts(ctx, page)
		if err != nil {
			return []Post{}, err
		}

		allPosts = append(allPosts, posts.Posts...)
		if posts.NextPage != nil {
			page = *posts.NextPage
		} else {
			break
		}
	}

	return allPosts, nil
}

func (c *Client) getPosts(ctx context.Context, page int) (*Posts, error) {
	path := fmt.Sprintf("teams/%s/posts", c.teamName)

	queries := map[string]string{
		"page":     strconv.Itoa(page),
		"per_page": "100",
		"include":  "comments",
	}

	req, err := c.newRequest(ctx, "GET", path, queries, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create new Request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get posts")
	}
	defer resp.Body.Close()

	fmt.Println("X-RateLimit-Remaining: ", resp.Header.Get("X-RateLimit-Remaining"))

	var posts Posts
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&posts); err != nil {
		return nil, errors.Wrap(err, "Falied to parse response")
	}

	return &posts, nil
}

func (c *Client) newRequest(ctx context.Context, method, requestPath string, queries map[string]string, body io.Reader) (*http.Request, error) {
	requestURL := *c.baseURL
	requestURL.Path = path.Join(c.baseURL.Path, requestPath)

	req, err := http.NewRequest(method, requestURL.String(), body)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for key, value := range queries {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	req = req.WithContext(ctx)
	bearerToken := fmt.Sprintf("Bearer %s", c.accessToken)
	req.Header.Set("Authorization", bearerToken)

	return req, nil
}
