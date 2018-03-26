package esa

import "time"

type Post struct {
	Number         int       `json:"number"`
	Name           string    `json:"name"`
	FullName       string    `json:"full_name"`
	Wip            bool      `json:"wip"`
	BodyMd         string    `json:"body_md"`
	BodyHTML       string    `json:"body_html"`
	CreatedAt      time.Time `json:"created_at"`
	Message        string    `json:"message"`
	URL            string    `json:"url"`
	UpdatedAt      time.Time `json:"updated_at"`
	Tags           []string  `json:"tags"`
	Category       string    `json:"category"`
	RevisionNumber int       `json:"revision_number"`
	CreatedBy      User      `json:"created_by"`
	UpdatedBy      User      `json:"updated_by"`

	Comments []Comment `json:"comments"`
}
