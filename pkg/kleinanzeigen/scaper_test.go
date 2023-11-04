package kleinanzeigen

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicScraperFunction(t *testing.T) {
	querier := NewQuerier(slog.Default())
	require.NotNil(t, querier)
	ads, err := querier.QueryAds(NewQuery("brompton", 500.0, 2000.0, WithLocality("Dortmund"), WithRadius(150)))
	require.NoError(t, err)
	assert.NotEmpty(t, ads)
}
