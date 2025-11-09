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

type Course struct {
	ID       int64 `json:"id"`
	MentorID int64 `json:"mentor_id"`
}

type Enrollment struct {
	IsEnrolled bool `json:"is_enrolled"`
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

// ✅ 1. User verify
func (c *Client) GetUser(id int64, token string) (*User, error) {
	url := fmt.Sprintf("%s/api/authentication/users/%d/", c.baseURL, id)

	// DEBUG LOGS
	fmt.Println("----- AUTH DEBUG -----")
	fmt.Println("REQUEST URL:", url)
	fmt.Println("TOKEN RAW:", token)
	fmt.Printf("TOKEN LENGTH: %d\n", len(token))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	fmt.Println("AUTH HEADER SENT:", req.Header.Get("Authorization"))

	resp, err := c.http.Do(req)
	if err != nil {
		fmt.Println("REQUEST ERROR:", err)
		return nil, err
	}
	fmt.Println("DJANGO RESPONSE STATUS:", resp.StatusCode)
	fmt.Println("-----------------------")

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("auth failed")
	}

	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	if u.ID != id {
		return nil, errors.New("token/user mismatch")
	}
	return &u, nil
}


// ✅ 2. Get course mentor
func (c *Client) GetCourse(id int64, token string) (int64, error) {
	url := fmt.Sprintf("%s/api/content/courses/%d/", c.baseURL, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, errors.New("course fetch failed")
	}

	var x Course
	if err := json.NewDecoder(resp.Body).Decode(&x); err != nil {
		return 0, err
	}
	return x.MentorID, nil
}

// ✅ 3. Check student access to course
func (c *Client) CheckEnrollment(courseID, userID int64, token string) (bool, error) {
	url := fmt.Sprintf("%s/api/content/courses/%d/students/%d/check/", c.baseURL, courseID, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, errors.New("enrollment check failed")
	}

	var e Enrollment
	if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
		return false, err
	}
	return e.IsEnrolled, nil
}
