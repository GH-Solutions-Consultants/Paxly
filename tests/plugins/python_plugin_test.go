// tests/plugins/python_plugin_test.go
package plugins_test

import (
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/GH-Solutions-Consultants/Paxly/core"
	"github.com/GH-Solutions-Consultants/Paxly/plugins/python"
	"github.com/stretchr/testify/mock" // Add this if you're using mock
)

// MockExecutor is a mock for executing commands.
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Run(cmd *core.Command) error {
	args := m.Called(cmd)
	return args.Error(0)
}

func (m *MockExecutor) Output(cmd *core.Command) ([]byte, error) {
	args := m.Called(cmd)
	return args.Get(0).([]byte), args.Error(1)
}

func TestPythonPlugin_Install_Success(t *testing.T) {
	// Initialize the mock executor
	mockExec := new(MockExecutor)

	// Setup expected commands and their outcomes
	mockExec.On("Run", mock.Anything).Return(nil).Twice() // create venv and install pipdeptree
	mockExec.On("Run", mock.Anything).Return(nil)            // pip install

	// Initialize the PythonPlugin with the mock executor
	plugin := python.NewPythonPlugin(mockExec)

	// Define dependencies to install
	deps := []core.Dependency{
		{
			Name:    "requests",
			Version: "^2.28",
		},
	}

	// Execute Install
	err := plugin.Install(deps)
	assert.NoError(t, err, "Expected Install to succeed")

	// Assert that Run was called
	mockExec.AssertExpectations(t)
}

func TestPythonPlugin_Install_Failure(t *testing.T) {
	// Initialize the mock executor
	mockExec := new(MockExecutor)

	// Setup expected commands and their outcomes
	mockExec.On("Run", mock.Anything).Return(errors.New("pip install failed"))

	// Initialize the PythonPlugin with the mock executor
	plugin := python.NewPythonPlugin(mockExec)

	// Define dependencies to install
	deps := []core.Dependency{
		{
			Name:    "requests",
			Version: "^2.28",
		},
	}

	// Execute Install
	err := plugin.Install(deps)
	assert.Error(t, err, "Expected Install to fail")

	// Assert that Run was called
	mockExec.AssertExpectations(t)
}

func TestPythonPlugin_List_Success(t *testing.T) {
	// Initialize the mock executor
	mockExec := new(MockExecutor)

	// Sample pip freeze output
	pipFreezeOutput := []byte("requests==2.28.1\nflask==1.1.2\n")

	// Setup expected commands and their outcomes
	mockExec.On("Output", mock.Anything).Return(pipFreezeOutput, nil)

	// Initialize the PythonPlugin with the mock executor
	plugin := python.NewPythonPlugin(mockExec)

	// Execute List
	deps, err := plugin.List()
	assert.NoError(t, err, "Expected List to succeed")
	assert.Len(t, deps, 2, "Expected two dependencies listed")

	assert.Equal(t, "requests", deps[0].Name)
	assert.Equal(t, "=2.28.1", deps[0].Version)

	assert.Equal(t, "flask", deps[1].Name)
	assert.Equal(t, "=1.1.2", deps[1].Version)

	// Assert that Output was called
	mockExec.AssertExpectations(t)
}

func TestPythonPlugin_List_Failure(t *testing.T) {
	// Initialize the mock executor
	mockExec := new(MockExecutor)

	// Setup expected commands and their outcomes
	mockExec.On("Output", mock.Anything).Return(nil, errors.New("pip freeze failed"))

	// Initialize the PythonPlugin with the mock executor
	plugin := python.NewPythonPlugin(mockExec)

	// Execute List
	deps, err := plugin.List()
	assert.Error(t, err, "Expected List to fail")
	assert.Nil(t, deps, "Expected no dependencies returned on failure")

	// Assert that Output was called
	mockExec.AssertExpectations(t)
}

func TestPythonPlugin_ListVersions_Success(t *testing.T) {
	// Initialize the mock executor
	mockExec := new(MockExecutor)

	// Mocked available versions
	availableVersions := []string{"2.25.0", "2.25.1", "2.26.0", "2.28.1"}

	// Setup expected commands and their outcomes
	mockExec.On("Run", mock.Anything).Return(nil) // pip install failed as expected
	mockExec.On("Run", mock.Anything).Return(nil)
	mockExec.On("Output", mock.Anything).Return([]byte{}, nil) // Simplified

	// Initialize the PythonPlugin with the mock executor
	plugin := python.NewPythonPlugin(mockExec)

	// Execute ListVersions
	versions, err := plugin.ListVersions("requests")
	assert.NoError(t, err, "Expected ListVersions to succeed")
	assert.Equal(t, availableVersions, versions, "Expected list of available versions")

	// Assert that Run was called
	mockExec.AssertExpectations(t)
}

func TestPythonPlugin_GetTransitiveDependencies_Success(t *testing.T) {
	// Initialize the mock executor
	mockExec := new(MockExecutor)

	// Sample pipdeptree JSON output
	pipDeptreeOutput := []byte(`[
		{
			"package": {"key": "requests", "name": "requests", "version": "2.28.1"},
			"dependencies": [
				{
					"package": {"key": "urllib3", "name": "urllib3", "version": "1.26.5"},
					"dependencies": []
				}
			]
		}
	]`)

	// Setup expected commands and their outcomes
	mockExec.On("Output", mock.Anything).Return(pipDeptreeOutput, nil)

	// Initialize the PythonPlugin with the mock executor
	plugin := python.NewPythonPlugin(mockExec)

	// Execute GetTransitiveDependencies
	transDeps, err := plugin.GetTransitiveDependencies("requests", "^2.28")
	assert.NoError(t, err, "Expected GetTransitiveDependencies to succeed")
	assert.Len(t, transDeps, 1, "Expected one transitive dependency")

	assert.Equal(t, "urllib3", transDeps[0].Name)
	assert.Equal(t, "=1.26.5", transDeps[0].Version)

	// Assert that Output was called
	mockExec.AssertExpectations(t)
}

func TestPythonPlugin_RunSecurityScan_NoVulnerabilities(t *testing.T) {
	// Initialize the mock executor
	mockExec := new(MockExecutor)

	// Sample safety check output with no vulnerabilities
	safetyOutput := []byte("[]")

	// Setup expected commands and their outcomes
	mockExec.On("Run", mock.Anything).Return(nil).Twice() // pip install safety and run safety check
	mockExec.On("Output", mock.Anything).Return(safetyOutput, nil)

	// Initialize the PythonPlugin with the mock executor
	plugin := python.NewPythonPlugin(mockExec)

	// Execute RunSecurityScan
	err := plugin.RunSecurityScan()
	assert.NoError(t, err, "Expected RunSecurityScan to pass with no vulnerabilities")

	// Assert that Run and Output were called
	mockExec.AssertExpectations(t)
}

func TestPythonPlugin_RunSecurityScan_WithVulnerabilities(t *testing.T) {
	// Initialize the mock executor
	mockExec := new(MockExecutor)

	// Sample safety check output with vulnerabilities
	safetyOutput := []byte(`[{"package": "requests", "vulnerability": "CVE-XXXX-XXXX"}]`)

	// Setup expected commands and their outcomes
	mockExec.On("Run", mock.Anything).Return(nil).Twice() // pip install safety and run safety check
	mockExec.On("Output", mock.Anything).Return(safetyOutput, nil)

	// Initialize the PythonPlugin with the mock executor
	plugin := python.NewPythonPlugin(mockExec)

	// Execute RunSecurityScan
	err := plugin.RunSecurityScan()
	assert.NoError(t, err, "Expected RunSecurityScan to complete even with vulnerabilities")

	// Assert that Run and Output were called
	mockExec.AssertExpectations(t)
}

func TestPythonPlugin_RunSecurityScan_Failure(t *testing.T) {
	// Initialize the mock executor
	mockExec := new(MockExecutor)

	// Setup expected commands and their outcomes
	mockExec.On("Run", mock.Anything).Return(errors.New("safety install failed"))

	// Initialize the PythonPlugin with the mock executor
	plugin := python.NewPythonPlugin(mockExec)

	// Execute RunSecurityScan
	err := plugin.RunSecurityScan()
	assert.Error(t, err, "Expected RunSecurityScan to fail due to safety install error")

	// Assert that Run was called
	mockExec.AssertExpectations(t)
}
