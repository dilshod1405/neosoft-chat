package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type User struct {
	ID       int64 `json:"id"`
	IsMentor bool  `json:"is_mentor"`
}

type Lesson struct {
	ID        int64 `json:"id"`
	CourseID  int64 `json:"course_id"`
	TeacherID int64 `json:"teacher_id"`
}

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) GetUser(id int64, token string) (*User, error) {
	url := fmt.Sprintf("%s/api/content/chat-service/users/%d/", c.baseURL, id)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("auth failed")
	}

	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}

	return &u, nil
}

func (c *Client) GetLesson(id int64, token string) (*Lesson, error) {
	url := fmt.Sprintf("%s/api/content/chat-service/lesson/%d/detail/", c.baseURL, id)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("lesson fetch failed")
	}

	var l Lesson
	if err := json.NewDecoder(resp.Body).Decode(&l); err != nil {
		return nil, err
	}

	return &l, nil
}
