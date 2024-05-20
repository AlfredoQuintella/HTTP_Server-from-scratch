package main

// net provides functions to work with networks.
// os provides an interface to interact with the operating system, such as terminating the program.
// bufio
// flag
// io
// path/filepath
// strconv
import (
	"fmt"
	"strings"
	"net"
	"os"
	"bufio"
	"flag"
	"io"
	"path/filepath"
	"strconv"
)

func main() {
	fmt.Println("Logs from your program")

	// Parse the directory flag.
	dirPtr := flag.String("directory", ".", "the directory to serve files from")
	flag.Parse()

	// Creating a TCP server that listens to the 4221 port in all network
	// interfaces. The function net.Listen return a listener ('1) and a possible error ('err').
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221:", err)
		os.Exit(1)
	}
	defer l.Close()

	for {
		// Accept incoming connections.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		// The function handleConnection is called in a goroutine. A goroutine is a light
		// thread maneged in the Go runtime.
		go handleConnection(conn, *dirPtr)
	}
}

// handleConnection processes an incoming connection and routes it based on the request.
func handleConnection(conn net.Conn, directory string) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	parts := strings.Fields(requestLine)
	if len(parts) < 3 {
		fmt.Println("Invalid request line")
		return
	}

	method := parts[0]
	urlPath := parts[1]

	// Read headers.
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
		}
	}

	if method == "POST" && strings.HasPrefix(urlPath, "/files/") {
		handlePostRequest(conn, reader, directory, urlPath, headers)
		return
	}

	if urlPath == "/user-agent" {
		handleUserAgent(conn, headers)
		return
	}

	if strings.HasPrefix(urlPath, "/files/") {
		handleFileRequest(conn, directory, urlPath)
		return
	}

	if strings.HasPrefix(urlPath, "/echo/") {
		handleEcho(conn, urlPath)
		return
	}

	if urlPath == "/" {
		handleRoot(conn)
		return
	}

	handleNotFound(conn)
}

// handlePostRequest processes POST requests to save files to the server.
func handlePostRequest(conn net.Conn, reader *bufio.Reader, directory, urlPath string, headers map[string]string) {
	filename := strings.TrimPrefix(urlPath, "/files/")
	filePath := filepath.Join(directory, filename)

	// Retrieve Content-Length from headers.
	contentLengthStr, ok := headers["Content-Length"]
	if !ok {
		fmt.Println("Missing Content-Length header")
		response := "HTTP/1.1 411 Length Required\r\n\r\n"
		conn.Write([]byte(response))
		return
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		fmt.Println("Invalid Content-Length:", err)
		response := "HTTP/1.1 400 Bad Request\r\n\r\n"
		conn.Write([]byte(response))
		return
	}

	// Read the body based on Content-Length.
	body := make([]byte, contentLength)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		fmt.Println("Error reading request body:", err)
		response := "HTTP/1.1 400 Bad Request\r\n\r\n"
		conn.Write([]byte(response))
		return
	}

	// Write the file to the specified directory.
	err = os.WriteFile(filePath, body, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		response := "HTTP/1.1 500 Internal Server Error\r\n\r\n"
		conn.Write([]byte(response))
		return
	}

	// Send a 201 Created response.
	response := "HTTP/1.1 201 Created\r\n\r\n"
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

// handleUserAgent responds with the User-Agent header sent by the client.
func handleUserAgent(conn net.Conn, headers map[string]string) {
	userAgent := headers["User-Agent"]
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)

	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

// handleFileRequest responds with the contents of the requested file.
func handleFileRequest(conn net.Conn, directory, urlPath string) {
	filename := strings.TrimPrefix(urlPath, "/files/")
	filePath := filepath.Join(directory, filename)

	// Read the file from the directory.
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		// File not found.
		response := "HTTP/1.1 404 Not Found\r\n\r\n"
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response:", err)
		}
		return
	}

	// Send the file content as the response.
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", len(fileContent))
	_, err = conn.Write(append([]byte(response), fileContent...))
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

// handleEcho responds with the text following /echo/ in the URL.
func handleEcho(conn net.Conn, urlPath string) {
	echoString := strings.TrimPrefix(urlPath, "/echo/")
	responseBody := echoString
	contentLength := len(responseBody)
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, responseBody)

	// Sending the response HTTP for the client.
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

// handleRoot responds with a basic 200 OK response for the root path.
func handleRoot(conn net.Conn) {
	response := "HTTP/1.1 200 OK\r\n\r\n"

	// Sending the response HTTP for the client.
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

// handleNotFound responds with a 404 Not Found status for unrecognized paths.
func handleNotFound(conn net.Conn) {
	response := "HTTP/1.1 404 Not Found\r\n\r\n"

	// Sending the response HTTP for the client.
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}
