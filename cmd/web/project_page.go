package main

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/gocasters/rankr/adapter/leaderboardstat"
	"github.com/gocasters/rankr/adapter/project"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const projectPagePath = "/projects/:id"

func registerProjectPage(router *echo.Echo, projects *project.HTTPClient, scores *leaderboardstat.HTTPClient) {
	router.GET(projectPagePath, func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil || id == uuid.Nil {
			return echo.NewHTTPError(http.StatusBadRequest, "شناسهٔ پروژه نامعتبر است.")
		}
		ctx := c.Request().Context()
		repo, found, err := projects.GetPublicProject(ctx, id)
		if err != nil {
			c.Logger().Error(err)
			return echo.ErrBadGateway
		}
		if !found {
			return echo.ErrNotFound
		}
		repoID, err := strconv.ParseUint(repo.RepositoryID, 10, 64)
		if err != nil || repoID == 0 {
			return echo.ErrBadGateway
		}

		result, err := scores.GetPublicLeaderboard(ctx, repoID, 10, 0)
		view := leaderboardView{ProjectID: strconv.FormatUint(repoID, 10), Project: &repo, Total: result.Total, Rows: result.Rows}
		status := http.StatusOK
		if err != nil {
			c.Logger().Error(err)
			view.Error = true
			status = http.StatusBadGateway
		}
		var output bytes.Buffer
		if err := renderLeaderboard(&output, view); err != nil {
			return err
		}
		return c.HTML(status, output.String())
	})
}
