package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/projectapp/constant"
	"github.com/gocasters/rankr/projectapp/service/project"
	vcs "github.com/gocasters/rankr/projectapp/service/versioncontrollersystemproject"
	"github.com/stretchr/testify/require"
)

const projectID = "9a0a1c5c-38c2-4717-b250-54515830b35f"

type publicProjectStore struct {
	project.Repository
	err error
}

func (s publicProjectStore) FindByID(_ context.Context, id string) (*project.ProjectEntity, error) {
	provider, repoID := constant.VcsProviderGitHub, "5001"
	return &project.ProjectEntity{ID: id, GitRepoID: &repoID, RepoProvider: &provider, Status: constant.ProjectStatusActive}, s.err
}

type publicRepoStore struct {
	vcs.Repository
	visibility constant.VcsVisibility
}

func (s publicRepoStore) FindByProviderID(_ context.Context, provider constant.VcsProvider, repoID, id string) (*vcs.VersionControllerSystemProjectEntity, error) {
	if id != projectID || repoID != "5001" || provider != constant.VcsProviderGitHub || s.visibility == "" {
		return nil, constant.ErrNotFound
	}
	return &vcs.VersionControllerSystemProjectEntity{
		Provider: provider, ProviderRepoID: repoID, Owner: "gocasters", Name: "rankr",
		Visibility: s.visibility, RemoteURL: "https://example.com/internal",
	}, nil
}

func TestPublicProjectRoute(t *testing.T) {
	for _, tt := range []struct {
		name       string
		id         string
		visibility constant.VcsVisibility
		projectErr error
		status     int
	}{
		{"public", projectID, constant.VcsVisibilityPublic, nil, http.StatusOK},
		{"private", projectID, constant.VcsVisibilityPrivate, nil, http.StatusNotFound},
		{"internal", projectID, constant.VcsVisibilityInternal, nil, http.StatusNotFound},
		{"unlinked", projectID, "", nil, http.StatusNotFound},
		{"missing", projectID, "", constant.ErrNotFound, http.StatusNotFound},
		{"invalid ID", "invalid", "", nil, http.StatusBadRequest},
		{"store error", projectID, "", errors.New("database unavailable"), http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			projects := project.NewService(publicProjectStore{err: tt.projectErr}, nil, logger)
			repos := vcs.NewService(publicRepoStore{visibility: tt.visibility}, nil, logger)
			httpServer, err := httpserver.New(httpserver.Config{
				Port:        8084,
				PublicPaths: []string{"/v1/projects/public/:id"},
			})
			require.NoError(t, err)
			server := New(httpServer, NewHandler(projects, repos, logger), logger, projects, repos)
			server.RegisterRoutes()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/projects/public/"+tt.id, nil)
			httpServer.GetRouter().ServeHTTP(rec, req)
			require.Equal(t, tt.status, rec.Code)
			if tt.status == http.StatusOK {
				require.JSONEq(t, `{"repository_id":"5001","name":"rankr","owner":"gocasters","provider":"GITHUB","status":"ACTIVE"}`, rec.Body.String())
			} else {
				require.NotContains(t, rec.Body.String(), "gocasters")
				require.NotContains(t, rec.Body.String(), "database unavailable")
			}
		})
	}
}
