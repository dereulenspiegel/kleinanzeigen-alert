package kleinanzeigenalert

import "time"

type Query interface {
	Term() string
	MaxPrice() float64
	MinPrice() float64
}

type Ad interface {
	ID() string
	CreatedAt() time.Time
	Description() string
	Title() string
	ImageURL() string
	Price() float64
	Link() string
}

type Querier[Q Query, A Ad] interface {
	QueryAds(query Q) ([]A, error)
	Name() string
}

type QueryUpdater[Q Query] func(Q)
type QueryOpt[Q Query] func(any) QueryUpdater[Q]
