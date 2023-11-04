package kleinanzeigen

import "time"

// Ad that is beeing stored
type Ad struct {
	id          string
	location    *string
	description string
	title       string
	imageUrl    string
	price       float64
	link        string
	createdAt   time.Time
}

func (a *Ad) ID() string {
	return a.id
}

func (a *Ad) Description() string {
	return a.description
}

func (a *Ad) Title() string {
	return a.title
}

func (a *Ad) ImageURL() string {
	return a.imageUrl
}

func (a *Ad) Price() float64 {
	return a.price
}

func (a *Ad) Link() string {
	return a.link
}

func (a *Ad) CreatedAt() time.Time {
	return a.createdAt
}
