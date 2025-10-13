# ONQL Project Index

Welcome to the Onqlbgfinal project! This document provides an overview of the project structure to help you navigate the codebase and contribute effectively.

## Project Structure

This project is organized into several Go packages, each with a specific responsibility.

### Core Components

*   **`engine`**: This package is responsible for managing the disk storage engine and interacting with NATS for messaging.
*   **`storemanager`**: This package handles low-level database operations. It creates instances of the engine but does not perform data validation.
*   **`database`**: This package provides a higher-level interface for database operations. It calls the `storemanager` APIs and enforces data validation rules.
*   **`dsl`**: This package implements the ONQL (Object Notation Query Language). It includes the lexer, parser, planner, and evaluator. The planner uses an "ONQL assembly" format, which represents every operation in a table-like structure.
*   **`server`**: This package contains the TCP server that clients connect to.
*   **`router`**: This package is responsible for routing incoming requests to the correct extension or internal service.
*   **`extensions`**: This folder contains core extensions that enhance the functionality of the database. These extensions communicate with the `router` via NATS. Examples include extensions for data import/export.

### Other Important Directories

*   **`cmd/server`**: The main application entry point.
*   **`config`**: Project configuration files.
*   **`docker`**: Docker-related files for building and deploying the application.
*   **`build`**: Contains build-related scripts and files.
*   **`logs`**: Directory for log files.
*   **`utils`**: Utility functions used across the project.

## Contributing

We welcome community contributions! There are many ways to contribute to the Onqlbgfinal project. Here are a few areas where you can help:

*   **DSL Enhancements**:
    *   Solve bugs in the ONQL DSL.
    *   Enhance existing aggregate functions or add new ones.
    *   Implement new features in the DSL.
*   **DBMS Core**:
    *   Solve bugs in the database management system.
    *   Fix issues in the storage engine.
*   **Plugins**:
    *   Develop new plugins to extend the database's functionality.
    *   Enhance existing plugins.
*   **Libraries and ORMs**:
    *   Create client libraries or ORMs for different programming languages.
*   **Documentation**:
    *   Improve the project documentation to make it clearer and more comprehensive.

If you're interested in contributing, please take a look at the following:

1.  **Issues**: Check the [issue tracker](.github/ISSUE_TEMPLATE) for bugs, feature requests, or other tasks.
2.  **Code Style**: Please follow the existing code style and conventions.
3.  **Pull Requests**: When submitting a pull request, please provide a clear description of the changes and why they are needed.

Thank you for your interest in Onqlbgfinal!

