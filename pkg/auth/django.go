package auth

import (
	// "encoding/json"
	// "errors"
	// "fmt"
	"net/http"
	"time"
)

type User struct {
	ID       int64 `json:"id"`
	IsMentor bool  `json:"is_mentor"`
}

type Course struct {
	ID       int64 `json:"id"`
	MentorID int64 `json:"mentor_id"`   // sizdagi JSON nomiga moslang
	Teacher  int64 `json:"teacher_id"`  // fallback
}

type Client struct {
	base string
	http *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		base: baseURL,
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

// func (c *Client) GetUser(id int64, token string) (*User, error) {
// 	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/authentication/users/%d/", c.base, id), nil)
// 	req.Header.Set("Authorization", "Bearer "+token)
// 	resp, err := c.http.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()
// 	if resp.StatusCode != 200 {
// 		return nil, errors.New("auth failed")
// 	}
// 	var u User
// 	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
// 		return nil, err
// 	}
// 	if u.ID != id {
// 		return nil, errors.New("token/user mismatch")
// 	}
// 	return &u, nil
// }

// func (c *Client) GetCourse(id int64, token string) (int64, error) {
// 	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/education/courses/%d/", c.base, id), nil)
// 	req.Header.Set("Authorization", "Bearer "+token)
// 	resp, err := c.http.Do(req)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer resp.Body.Close()
// 	if resp.StatusCode != 200 {
// 		return 0, errors.New("course fetch failed")
// 	}
// 	var x Course
// 	if err := json.NewDecoder(resp.Body).Decode(&x); err != nil {
// 		return 0, err
// 	}
// 	if x.MentorID != 0 {
// 		return x.MentorID, nil
// 	}
// 	if x.Teacher != 0 {
// 		return x.Teacher, nil
// 	}
// 	return 0, errors.New("mentor not found in course")
// }



func (c *Client) GetUser(id int64, token string) (*User, error) {
    return &User{ID: id, IsMentor: id < 1000}, nil // id < 1000 → mentor sifatida faraz qilamiz
}

func (c *Client) GetCourse(id int64, token string) (int64, error) {
    return 777, nil // har doim 777 mentor deb qaytaramiz
}
