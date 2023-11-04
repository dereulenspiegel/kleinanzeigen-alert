package kleinanzeigen

import kleinanzeigenalert "github.com/dereulenspiegel/kleinanzeigen-alert"

// Query that is beeing sored
type Query struct {
	term     string
	radius   int
	city     int
	cityName string
	maxPrice float64
	minPrice float64
}

func (q *Query) Term() string {
	return q.term
}

func (q *Query) MaxPrice() float64 {
	return q.maxPrice
}

func (q *Query) MinPrice() float64 {
	return q.minPrice
}

func WithLocality(city string) kleinanzeigenalert.QueryUpdater[*Query] {
	return func(q *Query) {
		q.cityName = city
	}
}

func WithRadius(radius int) kleinanzeigenalert.QueryUpdater[*Query] {
	return func(q *Query) {
		q.radius = radius
	}
}

func NewQuery(term string, minPrice, maxPrice float64, queryOpts ...kleinanzeigenalert.QueryUpdater[*Query]) *Query {
	q := &Query{
		term:     term,
		maxPrice: maxPrice,
		minPrice: minPrice,
	}
	for _, opt := range queryOpts {
		opt(q)
	}
	return q
}
