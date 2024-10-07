// core/plugin_registry.go
package core

import (
    "fmt"
    "sync"
)

// PluginAPIVersion defines the current version of the plugin API
const PluginAPIVersion = "1.0"

// PackageManagerPlugin defines the interface that all language-specific plugins must implement.
type PackageManagerPlugin interface {
    APIVersion() string
    Language() string
    Initialize(config Config) error
    Install(deps []Dependency) error
    Update(deps []Dependency) error
    Remove(dep Dependency) error
    List() ([]Dependency, error)
    ListVersions(depName string) ([]string, error)
    GetTransitiveDependencies(depName, version string) ([]Dependency, error)
    GetVulnerabilities() ([]SecurityVulnerability, error)
    Cleanup() error
}

// PluginRegistry manages all registered plugins.
type PluginRegistry struct {
    plugins map[string]PackageManagerPlugin
    mu      sync.RWMutex
}

// NewPluginRegistry creates a new PluginRegistry instance.
func NewPluginRegistry() *PluginRegistry {
    return &PluginRegistry{
        plugins: make(map[string]PackageManagerPlugin),
    }
}

// RegisterPlugin registers a new plugin.
func (pr *PluginRegistry) RegisterPlugin(lang string, plugin PackageManagerPlugin) error {
    pr.mu.Lock()
    defer pr.mu.Unlock()

    if plugin.APIVersion() != PluginAPIVersion {
        return fmt.Errorf("plugin %s has incompatible API version: got %s, expected %s",
            plugin.Language(), plugin.APIVersion(), PluginAPIVersion)
    }

    if _, exists := pr.plugins[lang]; exists {
        return fmt.Errorf("plugin for language '%s' is already registered", lang)
    }

    pr.plugins[lang] = plugin
    return nil
}

// GetPlugin retrieves a plugin by language.
func (pr *PluginRegistry) GetPlugin(lang string) (PackageManagerPlugin, bool) {
    pr.mu.RLock()
    defer pr.mu.RUnlock()

    plugin, exists := pr.plugins[lang]
    return plugin, exists
}

// GetAllPlugins retrieves all registered plugins.
func (pr *PluginRegistry) GetAllPlugins() map[string]PackageManagerPlugin {
    pr.mu.RLock()
    defer pr.mu.RUnlock()

    // Return a copy to prevent external modification
    pluginsCopy := make(map[string]PackageManagerPlugin)
    for lang, plugin := range pr.plugins {
        pluginsCopy[lang] = plugin
    }
    return pluginsCopy
}

// Singleton pattern for PluginRegistry
var (
    pluginRegistryInstance *PluginRegistry
    once                   sync.Once
)

// GetPluginRegistry returns the singleton instance of PluginRegistry.
func GetPluginRegistry() *PluginRegistry {
    once.Do(func() {
        pluginRegistryInstance = NewPluginRegistry()
    })
    return pluginRegistryInstance
}
