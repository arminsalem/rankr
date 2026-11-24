package http

import (
	"errors"
	"net/http"

	"github.com/gocasters/rankr/projectapp/constant"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h Handler) getPublicProject(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil || id == uuid.Nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project ID")
	}

	ctx := c.Request().Context()
	project, err := h.projectService.GetProject(ctx, id.String())
	if err != nil {
		return h.publicProjectError(err)
	}
	if project.GitRepoID == nil || project.RepoProvider == nil || *project.RepoProvider != constant.VcsProviderGitHub {
		return echo.ErrNotFound
	}

	repo, err := h.versionControllerSystemProjectService.GetVcsRepoByProviderID(ctx, *project.RepoProvider, *project.GitRepoID, project.ID)
	if err != nil {
		return h.publicProjectError(err)
	}
	if repo.Visibility != constant.VcsVisibilityPublic {
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, echo.Map{
		"repository_id": repo.ProviderRepoID,
		"name":          repo.Name,
		"owner":         repo.Owner,
		"provider":      repo.Provider,
		"status":        project.Status,
	})
}

func (h Handler) publicProjectError(err error) error {
	if errors.Is(err, constant.ErrNotFound) {
		return echo.ErrNotFound
	}
	h.logger.Error("failed to read public project", "error", err)
	return echo.ErrInternalServerError
}
