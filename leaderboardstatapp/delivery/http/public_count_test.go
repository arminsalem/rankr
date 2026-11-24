package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/cachemanager"
	"github.com/gocasters/rankr/pkg/logger"
	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type contributorCountCache struct {
	leaderboardstat.RedisLeaderboardRepository
}

func (contributorCountCache) GetPublicLeaderboardPaginated(context.Context, types.ID, int32, int32) ([]leaderboardstat.UserScoreEntry, int64, error) {
	return []leaderboardstat.UserScoreEntry{{UserID: 101, Score: 12}}, 21, nil
}

func TestPublicLeaderboardReturnsCountAcrossPages(t *testing.T) {
	require.NoError(t, logger.Init(logger.Config{Level: "error"}))
	service := leaderboardstat.NewService(nil, leaderboardstat.Validator{}, cachemanager.CacheManager{}, contributorCountCache{}, nil, nil, nil)
	router := echo.New()
	router.GET("/v1/leaderboard/public/:project_id", NewHandler(service).GetPublicLeaderboard)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/leaderboard/public/5001?page_size=1", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"project_id":"5001","total":21,"rows":[{"rank":1,"user_id":101,"score":12}]}`, rec.Body.String())
}
