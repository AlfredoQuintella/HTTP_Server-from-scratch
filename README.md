# HTTP_Server-from-scratch

This project implements a simple HTTP server from scratch in Go.

## Features

- Serves files from a specified directory.
- Handles `POST` requests to save files.
- Responds with `User-Agent` header.
- Echoes text in URL.
- Returns `404 Not Found` for unrecognized paths.

## Usage

1. Build the project:

    ```sh
    go build -o your_server
    ```

2. Run the server:

    ```sh
    ./your_server --directory /path/to/your/test/directory
    ```

3. Test the server using `curl`:

    - Create a file:

        ```sh
        curl -v -X POST http://localhost:4221/files/testfile.txt -H "Content-Length: 27" -d 'This is the content of the file.'
        ```

    - Get the file content:

        ```sh
        curl -v http://localhost:4221/files/testfile.txt
        ```

    - Get the `User-Agent`:

        ```sh
        curl -v http://localhost:4221/user-agent
        ```

    - Echo a string:

        ```sh
        curl -v http://localhost:4221/echo/hello
        ```
