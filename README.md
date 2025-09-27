# majoo-tech-assesment

# Simple Blog API

This project is a simple, yet robust, RESTful API for a blog platform, built with Go. It features user registration, authentication using JWT, and full CRUD (Create, Read, Update, Delete) functionality for posts and comments. The entire application is containerized using Docker for easy setup and deployment.

## Table of Contents

- [Architecture](#architecture)
- [Technology Choices](#technology-choices)
- [Setup and Running](#setup-and-running)
- [API Endpoints](#api-endpoints)
- [Testing](#testing)
- [Known Limitations & Future Improvements](#known-limitations--future-improvements)

## Architecture

The application follows a classic layered (or clean) architecture to ensure separation of concerns, making it modular, scalable, and easy to maintain.

1.  **Handler Layer (`/internal/handler`)**: This layer is responsible for handling incoming HTTP requests. It parses request bodies, validates input, and calls the appropriate service method. It is the entry point for all API interactions.

2.  **Service Layer (`/internal/service`)**: This layer contains the core business logic of the application. It orchestrates operations, performs complex validations, and coordinates between different repositories. For example, the `UserService` handles the logic for hashing passwords and generating JWTs.

3.  **Repository Layer (`/internal/repository`)**: This layer is responsible for all communication with the database. It abstracts the database operations, providing a clean interface for the service layer to interact with data without knowing the underlying database implementation details.

4.  **Model Layer (`/internal/model`)**: This layer defines the core data structures (structs) used throughout the application, such as `User`, `Post`, and `Comment`.

### Containerization & Orchestration

-   **Docker**: The application and its database are fully containerized, ensuring a consistent and reproducible environment across different machines.
-   **Docker Compose**: `docker-compose.yml` is used to define and orchestrate the multi-container setup. It manages the `api` service, the `postgres` database service, and a one-off `migrate` service.
-   **Database Migrations**: The `migrate` service uses `golang-migrate` to automatically apply SQL schema migrations upon startup. This ensures the database schema is always in sync with the application's expectations. The `api` service waits for the migrations to complete successfully before starting.

## Technology Choices

| Technology          | Justification                                                                                                                                                           |
| ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Go (Golang)**     | Chosen for its high performance, excellent support for concurrency, simplicity, and strong standard library. It's ideal for building efficient and scalable backend APIs. |
| **Gin**             | A minimalistic and high-performance web framework for Go. Its speed and low overhead make it perfect for building REST APIs.                                              |
| **PostgreSQL**      | A powerful, open-source, and feature-rich relational database system known for its reliability, robustness, and performance.                                            |
| **Docker**          | Provides a consistent, isolated, and reproducible environment for development, testing, and deployment, eliminating "it works on my machine" problems.                    |
| **`golang-migrate`**| A powerful CLI tool for managing database schema migrations. It allows for version-controlled, automated schema changes, which is crucial for CI/CD pipelines.          |
| **`golang-jwt`**    | The standard library for creating and verifying JSON Web Tokens (JWTs), used here for stateless user authentication.                                                      |
| **`bcrypt`**        | The industry-standard library for securely hashing and salting user passwords. It's computationally expensive, which helps protect against brute-force attacks.          |
| **`gomock` / `testify`** | A powerful combination for unit testing in Go. `gomock` allows for the creation of mock implementations of interfaces (like our repositories), enabling isolated service-layer testing. `testify` provides a rich set of assertion tools. |

## Setup and Running

### Prerequisites

-   Docker
-   Docker Compose

### Instructions

1.  **Clone the repository:**
    ```sh
    git clone <repository-url>
    cd simple-blog-api
    ```

2.  **Build and run the application:**
    ```sh
    docker compose up -d --build 
    ```

    This single command will:
    -   Build the Go application and `migrate` tool inside a Docker image.
    -   Start the PostgreSQL database container.
    -   Wait for the database to be healthy.
    -   Run the database migrations to set up the schema and seed initial data.
    -   Start the API server, which will be accessible on `http://localhost:8080`.

3.  **Stopping the application:**
    To stop the containers, press `Ctrl+C` in the terminal where `docker-compose` is running, or run:
    ```sh
    docker compose down
    ```

    To stop the containers and remove the database volume (for a complete reset):
    ```sh
    docker compose down -v
    ```

## API Endpoints

The base URL is `http://localhost:8080`.
You can check the detailed Swagger API Documentation on `http://localhost:8080/swagger/index.html`

| Endpoint                  | Method | Auth Required | Description                               |
| ------------------------- | ------ | ------------- | ----------------------------------------- |
| `/register`               | `POST` | No            | Registers a new user.                     |
| `/login`                  | `POST` | No            | Logs in a user and returns a JWT.         |
| `/posts`                  | `POST` | Yes           | Creates a new post.                       |
| `/posts`                  | `GET`  | No            | Retrieves a list of all posts.            |
| `/posts/{id}`             | `GET`  | No            | Retrieves a single post by its ID.        |
| `/posts/{id}`             | `PUT`  | Yes           | Updates a post. (Must be the post owner)  |
| `/posts/{id}`             | `DELETE`| Yes           | Deletes a post. (Must be the post owner)  |
| `/posts/{id}/comments`    | `POST` | Yes           | Adds a comment to a post.                 |
| `/comments/{id}`          | `PUT`  | Yes           | Updates a comment. (Must be the owner)    |
| `/comments/{id}`          | `DELETE`| Yes           | Deletes a comment. (Must be the owner)    |

**Authentication**: For protected endpoints, include the JWT in the `Authorization` header:
`Authorization: Bearer <your-jwt-token>`

## Testing

The project has a comprehensive suite of unit tests for the handler, service, and middleware layers. The repository layer is tested implicitly through the service layer tests, which use mock repositories.

### Running Tests

To run all tests, execute the following command from the project root:

```sh
go test ./...
```

### Test Coverage

To generate a test coverage report, run the following commands:

```sh
# Generate a coverage profile
go test -coverprofile=coverage.out ./...

# View the interactive HTML report in your browser
go tool cover -html=coverage.out
```

This will open a browser window showing which lines of code are covered by the tests, helping to identify untested logic.

## Known Limitations & Future Improvements

### Known Limitations

-   **No Pagination**: Endpoints that return lists (e.g., `GET /posts`) do not currently support pagination. This can lead to performance issues and large response bodies if the database contains many records.
-   **Simple Error Handling**: While functional, the error responses could be more structured, following a standard like RFC 7807 for Problem Details for HTTP APIs.
-   **No Request Logging**: There is no middleware for logging incoming HTTP requests, which is essential for debugging and monitoring in a production environment.
-   **No Role-Based Access Control (RBAC)**: The authentication system is simple and does not differentiate between user roles (e.g., `admin` vs. `user`). An admin cannot, for example, delete another user's post.

### Future Improvements

-   **Implement Pagination**: Add query parameters (`page`, `limit`) to list endpoints to allow clients to fetch data in chunks.
-   **Structured Logging**: Introduce a logging middleware using a library like `zerolog` or `zap` to log request details in a structured (JSON) format.
-   **Implement RBAC**: Enhance the JWT claims to include user roles and update the authentication middleware to check for these roles, allowing for more granular access control.
-   **Add Caching**: Introduce a caching layer (e.g., using Redis) for frequently accessed, non-critical data (like public posts) to reduce database load and improve response times.
-   **Configuration Management**: Move configuration from `docker-compose.yml` environment variables to a dedicated configuration management system or file (e.g., using Viper) for better flexibility across different environments.
-   **Add More Input Validation**: Implement more robust validation on request bodies using a library like `go-playground/validator`.