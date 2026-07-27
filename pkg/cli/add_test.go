package cli

import (
	"testing"

	"github.com/devports/devpt/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestAddCmdForceOverwritesPreservingRuntime(t *testing.T) {
	app, _, _ := newTestApp(t)
	pid := 1111
	require.NoError(t, app.registry.AddService(&models.ManagedService{
		Name: "svc", CWD: "/old", Command: "sleep 1", Ports: []int{3000}, LastPID: &pid,
	}))

	require.NoError(t, app.AddCmd("svc", "/new", "npm run dev", []int{8080}, true))

	got := app.registry.GetService("svc")
	require.NotNil(t, got)
	require.Equal(t, "/new", got.CWD)
	require.Equal(t, "npm run dev", got.Command)
	require.Equal(t, []int{8080}, got.Ports)
	require.NotNil(t, got.LastPID, "LastPID must be preserved on --force overwrite")
	require.Equal(t, 1111, *got.LastPID)
}

func TestAddCmdForceCreatesWhenAbsent(t *testing.T) {
	app, _, _ := newTestApp(t)
	require.NoError(t, app.AddCmd("brand", "/p", "go run .", []int{9000}, true))
	got := app.registry.GetService("brand")
	require.NotNil(t, got)
	require.False(t, got.CreatedAt.IsZero(), "CreatedAt should be set when force-creating")
}

func TestAddCmdRejectsExistingWithoutForce(t *testing.T) {
	app, _, _ := newTestApp(t)
	require.NoError(t, app.AddCmd("svc", "/p", "sleep 1", []int{3000}, false))
	err := app.AddCmd("svc", "/p", "sleep 2", []int{4000}, false)
	require.Error(t, err, "non-force add of existing name must still be rejected")
}

func TestAddCmdRejectsEmptyNameAndBadCommand(t *testing.T) {
	app, _, _ := newTestApp(t)
	require.Error(t, app.AddCmd("  ", "/p", "sleep 1", nil, false))
	require.Error(t, app.AddCmd("svc", "/p", "a && b", nil, false))
}
