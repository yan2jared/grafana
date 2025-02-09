package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/dashboardimport"
	"github.com/grafana/grafana/pkg/web/webtest"
	"github.com/stretchr/testify/require"
)

func TestImportDashboardAPI(t *testing.T) {
	t.Run("Quota not reached, schema loader service disabled", func(t *testing.T) {
		importDashboardServiceCalled := false
		service := &serviceMock{
			importDashboardFunc: func(ctx context.Context, req *dashboardimport.ImportDashboardRequest) (*dashboardimport.ImportDashboardResponse, error) {
				importDashboardServiceCalled = true
				return nil, nil
			},
		}

		schemaLoaderServiceCalled := false
		schemaLoaderService := &schemaLoaderServiceMock{
			dashboardApplyDefaultsFunc: func(input *simplejson.Json) (*simplejson.Json, error) {
				schemaLoaderServiceCalled = true
				return input, nil
			},
		}

		importDashboardAPI := New(service, quotaServiceFunc(quotaNotReached), schemaLoaderService, nil)
		routeRegister := routing.NewRouteRegister()
		importDashboardAPI.RegisterAPIEndpoints(routeRegister)
		s := webtest.NewServer(t, routeRegister)

		t.Run("Not signed in should return 404", func(t *testing.T) {
			cmd := &dashboardimport.ImportDashboardRequest{}
			jsonBytes, err := json.Marshal(cmd)
			require.NoError(t, err)
			req := s.NewRequest(http.MethodPost, "/api/dashboards/import", bytes.NewReader(jsonBytes))
			resp, err := s.Send(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("Signed in, empty plugin id and dashboard model empty should return error", func(t *testing.T) {
			cmd := &dashboardimport.ImportDashboardRequest{
				PluginId:  "",
				Dashboard: nil,
			}
			jsonBytes, err := json.Marshal(cmd)
			require.NoError(t, err)
			req := s.NewRequest(http.MethodPost, "/api/dashboards/import", bytes.NewReader(jsonBytes))
			webtest.RequestWithSignedInUser(req, &models.SignedInUser{
				UserId: 1,
			})
			resp, err := s.Send(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		})

		t.Run("Signed in, dashboard model set should call import dashboard service", func(t *testing.T) {
			cmd := &dashboardimport.ImportDashboardRequest{
				Dashboard: simplejson.New(),
			}
			jsonBytes, err := json.Marshal(cmd)
			require.NoError(t, err)
			req := s.NewRequest(http.MethodPost, "/api/dashboards/import", bytes.NewReader(jsonBytes))
			webtest.RequestWithSignedInUser(req, &models.SignedInUser{
				UserId: 1,
			})
			resp, err := s.Send(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.True(t, importDashboardServiceCalled)
		})

		t.Run("Signed in, dashboard model set, trimdefaults enabled should not call schema loader service", func(t *testing.T) {
			cmd := &dashboardimport.ImportDashboardRequest{
				Dashboard: simplejson.New(),
			}
			jsonBytes, err := json.Marshal(cmd)
			require.NoError(t, err)
			req := s.NewRequest(http.MethodPost, "/api/dashboards/import?trimdefaults=true", bytes.NewReader(jsonBytes))
			webtest.RequestWithSignedInUser(req, &models.SignedInUser{
				UserId: 1,
			})
			resp, err := s.Send(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.False(t, schemaLoaderServiceCalled)
			require.True(t, importDashboardServiceCalled)
		})
	})

	t.Run("Quota not reached, schema loader service enabled", func(t *testing.T) {
		importDashboardServiceCalled := false
		service := &serviceMock{
			importDashboardFunc: func(ctx context.Context, req *dashboardimport.ImportDashboardRequest) (*dashboardimport.ImportDashboardResponse, error) {
				importDashboardServiceCalled = true
				return nil, nil
			},
		}

		schemaLoaderServiceCalled := false
		schemaLoaderService := &schemaLoaderServiceMock{
			enabled: true,
			dashboardApplyDefaultsFunc: func(input *simplejson.Json) (*simplejson.Json, error) {
				schemaLoaderServiceCalled = true
				return input, nil
			},
		}

		importDashboardAPI := New(service, quotaServiceFunc(quotaNotReached), schemaLoaderService, nil)
		routeRegister := routing.NewRouteRegister()
		importDashboardAPI.RegisterAPIEndpoints(routeRegister)
		s := webtest.NewServer(t, routeRegister)

		t.Run("Signed in, dashboard model set, trimdefaults enabled should call schema loader service", func(t *testing.T) {
			cmd := &dashboardimport.ImportDashboardRequest{
				Dashboard: simplejson.New(),
			}
			jsonBytes, err := json.Marshal(cmd)
			require.NoError(t, err)
			req := s.NewRequest(http.MethodPost, "/api/dashboards/import?trimdefaults=true", bytes.NewReader(jsonBytes))
			webtest.RequestWithSignedInUser(req, &models.SignedInUser{
				UserId: 1,
			})
			resp, err := s.Send(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.True(t, schemaLoaderServiceCalled)
			require.True(t, importDashboardServiceCalled)
		})
	})

	t.Run("Quota reached", func(t *testing.T) {
		service := &serviceMock{}
		schemaLoaderService := &schemaLoaderServiceMock{}
		importDashboardAPI := New(service, quotaServiceFunc(quotaReached), schemaLoaderService, nil)

		routeRegister := routing.NewRouteRegister()
		importDashboardAPI.RegisterAPIEndpoints(routeRegister)
		s := webtest.NewServer(t, routeRegister)

		t.Run("Signed in, dashboard model set, should return 403 forbidden/quota reached", func(t *testing.T) {
			cmd := &dashboardimport.ImportDashboardRequest{
				Dashboard: simplejson.New(),
			}
			jsonBytes, err := json.Marshal(cmd)
			require.NoError(t, err)
			req := s.NewRequest(http.MethodPost, "/api/dashboards/import", bytes.NewReader(jsonBytes))
			webtest.RequestWithSignedInUser(req, &models.SignedInUser{
				UserId: 1,
			})
			resp, err := s.Send(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusForbidden, resp.StatusCode)
		})
	})
}

type serviceMock struct {
	importDashboardFunc func(ctx context.Context, req *dashboardimport.ImportDashboardRequest) (*dashboardimport.ImportDashboardResponse, error)
}

func (s *serviceMock) ImportDashboard(ctx context.Context, req *dashboardimport.ImportDashboardRequest) (*dashboardimport.ImportDashboardResponse, error) {
	if s.importDashboardFunc != nil {
		return s.importDashboardFunc(ctx, req)
	}

	return nil, nil
}

type schemaLoaderServiceMock struct {
	enabled                    bool
	dashboardApplyDefaultsFunc func(input *simplejson.Json) (*simplejson.Json, error)
}

func (s *schemaLoaderServiceMock) IsDisabled() bool {
	return !s.enabled
}

func (s *schemaLoaderServiceMock) DashboardApplyDefaults(input *simplejson.Json) (*simplejson.Json, error) {
	if s.dashboardApplyDefaultsFunc != nil {
		return s.dashboardApplyDefaultsFunc(input)
	}

	return input, nil
}

func quotaReached(c *models.ReqContext, target string) (bool, error) {
	return true, nil
}

func quotaNotReached(c *models.ReqContext, target string) (bool, error) {
	return false, nil
}
