
# Paxly - Multi-Language Dependency Management Made Easy

Paxly is a powerful, multi-language package manager that simplifies dependency management across various programming languages. With a unified interface, Paxly helps developers manage dependencies seamlessly for Python, JavaScript, Go, Rust, and more through an extensible plugin architecture.

🚀 **Features**

- **Multi-Language Support**: Manage dependencies for multiple programming languages, including Python, JavaScript, Go, and Rust.
- **Plugin Architecture**: Extend Paxly to support additional languages with ease. Develop and share your own plugins.
- **Environment-Specific Configurations**: Separate dependencies by environment (e.g., development, testing, production) for better project organization.
- **Recursive Dependency Resolution**: Automatically resolve transitive dependencies using semantic versioning to ensure compatibility.
- **Parallel Installations**: Boost installation speed with concurrent operations, reducing setup time.
- **Efficient Caching**: Cache downloaded packages to prevent redundant downloads and speed up future installations.
- **Security Integrations**: Identify vulnerabilities in your dependencies with built-in security scans.
- **Cross-Platform Compatibility**: Paxly runs smoothly on Windows, and Linux.
- **Comprehensive Logging**: Gain insights into Paxly's processes with detailed logging and customizable verbosity levels.
- **Interactive CLI Mode**: Use Paxly interactively via its built-in shell for efficient dependency management.
- **Dependency Graph Visualization**: Visualize dependency relationships in your projects with Graphviz integration (Not available yet).

![Recording 2024-10-14 211425 mp4](https://github.com/user-attachments/assets/c60f4a5d-097b-456f-a461-1b6b6e41c851)

🛠 **Installation**

### Prerequisites

To use Paxly, make sure you have the following tools installed on your system:

- **Go** (for building Paxly itself): [Go Downloads](https://golang.org/dl/)
- **Python** (python3 and pip) for Python dependency management
- **Node.js** (npm) for JavaScript projects
- **Rust** (cargo) for Rust dependency management (install via rustup)
- **Graphviz** for dependency graph visualization: [Graphviz Downloads](https://graphviz.org/download/)

### Clone the Repository

To get started with Paxly, clone the repository to your local machine:
```bash
git clone https://github.com/yourusername/paxly.git
cd paxly
```

### Building Paxly

Build Paxly using Go:

```bash
go build
```

### Adding Paxly to Your PATH

For easier access, consider adding Paxly to your system's PATH:

For **Windows**, add the Paxly directory to the system environment variable `PATH` via System Properties.

🎉 **Getting Started**

### Initializing a New Project

To initialize a new Paxly project, use the `init` command:

```bash
paxly init
```

This command creates a `paxly.yaml` file that stores project metadata and dependency information.

### Adding Dependencies

Add a dependency to your project using the `add` command:

```bash
paxly add python requests>=2.28
```

This command adds the Python package `requests` with a compatible version of 2.28 or higher to the project.

### Listing Dependencies

To see the list of dependencies in your project:

```bash
paxly list
```

Paxly displays all installed dependencies for each configured environment.

### Installing Dependencies

Install all the dependencies defined in your `paxly.yaml` file:

```bash
paxly install
```

Paxly automatically resolves and installs all required dependencies.

### Removing Dependencies

To remove a dependency from your project:

```bash
paxly remove python requests
```

### Updating Dependencies

Update a specific dependency to the latest compatible version:

```bash
paxly update python requests
```

### Visualizing Dependency Graphs

To visualize the dependency graph for your project, make sure Graphviz is installed and run:

```bash
paxly graph
```

This command generates a visual representation of your dependencies in `graph.png`.

🔌 **Extending Paxly with Plugins**

Paxly's plugin system allows you to extend support for additional programming languages or add new functionality. Each plugin must implement the `PackageManagerPlugin` interface.

### To add a custom plugin:

1. Implement the `PackageManagerPlugin` interface.
2. Register your plugin using `core.GetPluginRegistry().RegisterPlugin(...)`.
3. Rebuild Paxly to use your custom plugin.

🛡 **Security and Best Practices**

Paxly helps maintain the security of your projects by integrating with popular security tools:

- **Python**: Uses `safety` to check for vulnerabilities in Python dependencies.
- **JavaScript**: Uses `npm audit` to identify security issues.
- **Rust**: Uses `cargo audit` for vulnerability checks.

Run the security scan command to identify and address vulnerabilities:

```bash
paxly scan
```

🤝 **Contributing**

Paxly is an open-source project, and we welcome contributions! To contribute:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Submit a pull request.

Please see our `CONTRIBUTING.md` for more details.

📄 **License**

Paxly is licensed under the MIT License. See [LICENSE](LICENSE) for more information.

💬 **Get in Touch**

- **Issues**: Found a bug? Have a feature request? Submit an issue on our [GitHub Issues](https://github.com/yourusername/paxly/issues) page.
- **Community**: Join our community discussions on Discord (link coming soon).

📚 **Documentation**

For more details, see the full [Paxly Documentation](https://github.com/yourusername/paxly/wiki).

🏗 **Roadmap**

Paxly is continually evolving, and we're working on exciting new features:

- **Support for Additional Languages**: Add support for new languages based on community feedback.
- **GUI for Paxly**: Develop a desktop GUI application for users who prefer not to use the command line.
- **Improved Plugin Marketplace**: Allow plugin ratings, reviews, and one-click installation.

Stay tuned for updates and upcoming releases!

