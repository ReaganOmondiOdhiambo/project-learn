/*
 * Simple HTTP Server in C
 * 
 * This is a basic HTTP/1.1 server implementation that demonstrates:
 * - Socket programming (creating and binding sockets)
 * - TCP connections (listening and accepting clients)
 * - HTTP protocol basics (parsing requests, sending responses)
 * - Serving static files from the filesystem
 * 
 * Compile: gcc -o http_server http_server.c
 * Run: ./http_server
 * Test: Open http://localhost:8080 in your browser
 */

#include <stdio.h>      // For printf, fprintf, perror, FILE operations
#include <stdlib.h>     // For exit, malloc, free
#include <string.h>     // For string operations like strlen, strcpy, strstr
#include <unistd.h>     // For close, read, write
#include <sys/socket.h> // For socket functions (socket, bind, listen, accept)
#include <netinet/in.h> // For sockaddr_in structure
#include <sys/stat.h>   // For stat function (checking file existence)
#include <fcntl.h>      // For file operations (open)

// Configuration constants
#define PORT 8080           // Port number the server will listen on
#define BUFFER_SIZE 8192    // Size of buffer for reading requests
#define MAX_CONNECTIONS 10  // Maximum number of queued connections

/*
 * Function: get_content_type
 * --------------------------
 * Determines the MIME type based on file extension.
 * This tells the browser how to interpret the file.
 * 
 * filepath: The path to the file
 * 
 * returns: A string representing the MIME type
 */
const char* get_content_type(const char* filepath) {
    // Find the last dot in the filename to get the extension
    const char* dot = strrchr(filepath, '.');
    
    if (!dot) {
        return "text/plain"; // No extension found, default to plain text
    }
    
    // Compare the extension and return appropriate MIME type
    if (strcmp(dot, ".html") == 0) return "text/html";
    if (strcmp(dot, ".css") == 0) return "text/css";
    if (strcmp(dot, ".js") == 0) return "application/javascript";
    if (strcmp(dot, ".json") == 0) return "application/json";
    if (strcmp(dot, ".png") == 0) return "image/png";
    if (strcmp(dot, ".jpg") == 0 || strcmp(dot, ".jpeg") == 0) return "image/jpeg";
    if (strcmp(dot, ".gif") == 0) return "image/gif";
    if (strcmp(dot, ".svg") == 0) return "image/svg+xml";
    
    return "text/plain"; // Default fallback
}

/*
 * Function: send_response
 * -----------------------
 * Sends an HTTP response to the client.
 * 
 * client_socket: The socket descriptor for the connected client
 * status_code: HTTP status code (e.g., 200, 404, 500)
 * status_text: HTTP status text (e.g., "OK", "Not Found")
 * content_type: MIME type of the response body
 * body: The actual content to send
 * body_length: Length of the body in bytes
 */
void send_response(int client_socket, int status_code, const char* status_text, 
                   const char* content_type, const char* body, size_t body_length) {
    char header[BUFFER_SIZE];
    
    // Build the HTTP response header
    // Format: HTTP/1.1 [status_code] [status_text]\r\n
    //         Content-Type: [content_type]\r\n
    //         Content-Length: [body_length]\r\n
    //         Connection: close\r\n
    //         \r\n
    int header_length = snprintf(header, BUFFER_SIZE,
        "HTTP/1.1 %d %s\r\n"
        "Content-Type: %s\r\n"
        "Content-Length: %zu\r\n"
        "Connection: close\r\n"
        "\r\n",
        status_code, status_text, content_type, body_length
    );
    
    // Send the header to the client
    write(client_socket, header, header_length);
    
    // Send the body to the client (if there is one)
    if (body && body_length > 0) {
        write(client_socket, body, body_length);
    }
}

/*
 * Function: send_file
 * -------------------
 * Reads a file from disk and sends it as an HTTP response.
 * 
 * client_socket: The socket descriptor for the connected client
 * filepath: Path to the file to send
 */
void send_file(int client_socket, const char* filepath) {
    // Check if file exists using stat
    struct stat file_stat;
    if (stat(filepath, &file_stat) == -1) {
        // File doesn't exist, send 404 Not Found
        const char* not_found = "<html><body><h1>404 Not Found</h1></body></html>";
        send_response(client_socket, 404, "Not Found", "text/html", 
                     not_found, strlen(not_found));
        return;
    }
    
    // Open the file for reading in binary mode
    FILE* file = fopen(filepath, "rb");
    if (!file) {
        // Error opening file, send 500 Internal Server Error
        const char* error = "<html><body><h1>500 Internal Server Error</h1></body></html>";
        send_response(client_socket, 500, "Internal Server Error", "text/html", 
                     error, strlen(error));
        return;
    }
    
    // Allocate memory to hold the entire file
    char* file_content = malloc(file_stat.st_size);
    if (!file_content) {
        fclose(file);
        const char* error = "<html><body><h1>500 Internal Server Error</h1></body></html>";
        send_response(client_socket, 500, "Internal Server Error", "text/html", 
                     error, strlen(error));
        return;
    }
    
    // Read the entire file into memory
    fread(file_content, 1, file_stat.st_size, file);
    fclose(file);
    
    // Determine the content type based on file extension
    const char* content_type = get_content_type(filepath);
    
    // Send the file as a 200 OK response
    send_response(client_socket, 200, "OK", content_type, 
                 file_content, file_stat.st_size);
    
    // Free the allocated memory
    free(file_content);
}

/*
 * Function: handle_request
 * ------------------------
 * Parses the HTTP request and routes it appropriately.
 * 
 * client_socket: The socket descriptor for the connected client
 * request: The raw HTTP request string
 */
void handle_request(int client_socket, const char* request) {
    char method[16];    // HTTP method (GET, POST, etc.)
    char path[256];     // Requested path
    char protocol[16];  // HTTP protocol version
    
    // Parse the first line of the HTTP request
    // Format: METHOD /path HTTP/1.1
    sscanf(request, "%s %s %s", method, path, protocol);
    
    printf("Request: %s %s %s\n", method, path, protocol);
    
    // Only handle GET requests in this simple server
    if (strcmp(method, "GET") != 0) {
        const char* not_implemented = "<html><body><h1>501 Not Implemented</h1></body></html>";
        send_response(client_socket, 501, "Not Implemented", "text/html", 
                     not_implemented, strlen(not_implemented));
        return;
    }
    
    // Build the file path
    char filepath[512] = "."; // Start with current directory
    
    // If path is just "/", serve index.html
    if (strcmp(path, "/") == 0) {
        strcat(filepath, "/index.html");
    } else {
        // Otherwise, append the requested path
        strcat(filepath, path);
    }
    
    // Send the requested file
    send_file(client_socket, filepath);
}

/*
 * Function: main
 * --------------
 * Entry point of the server program.
 * Sets up the socket, binds to a port, and handles incoming connections.
 */
int main() {
    int server_socket;              // Socket descriptor for the server
    int client_socket;              // Socket descriptor for each client
    struct sockaddr_in server_addr; // Server address structure
    struct sockaddr_in client_addr; // Client address structure
    socklen_t client_addr_len;      // Length of client address structure
    char buffer[BUFFER_SIZE];       // Buffer for reading requests
    
    // Step 1: Create a socket
    // AF_INET = IPv4, SOCK_STREAM = TCP, 0 = default protocol
    server_socket = socket(AF_INET, SOCK_STREAM, 0);
    if (server_socket == -1) {
        perror("Failed to create socket");
        exit(EXIT_FAILURE);
    }
    
    // Step 2: Set socket options
    // SO_REUSEADDR allows reusing the address immediately after the server stops
    int opt = 1;
    if (setsockopt(server_socket, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) == -1) {
        perror("Failed to set socket options");
        close(server_socket);
        exit(EXIT_FAILURE);
    }
    
    // Step 3: Configure server address
    memset(&server_addr, 0, sizeof(server_addr)); // Zero out the structure
    server_addr.sin_family = AF_INET;              // IPv4
    server_addr.sin_addr.s_addr = INADDR_ANY;      // Accept connections on any network interface
    server_addr.sin_port = htons(PORT);            // Convert port to network byte order
    
    // Step 4: Bind the socket to the address and port
    if (bind(server_socket, (struct sockaddr*)&server_addr, sizeof(server_addr)) == -1) {
        perror("Failed to bind socket");
        close(server_socket);
        exit(EXIT_FAILURE);
    }
    
    // Step 5: Listen for incoming connections
    // MAX_CONNECTIONS is the backlog (max queued connections)
    if (listen(server_socket, MAX_CONNECTIONS) == -1) {
        perror("Failed to listen on socket");
        close(server_socket);
        exit(EXIT_FAILURE);
    }
    
    printf("HTTP Server is running on http://localhost:%d\n", PORT);
    printf("Press Ctrl+C to stop the server\n\n");
    
    // Step 6: Main server loop - accept and handle connections
    while (1) {
        client_addr_len = sizeof(client_addr);
        
        // Accept a new connection (this blocks until a client connects)
        client_socket = accept(server_socket, (struct sockaddr*)&client_addr, &client_addr_len);
        if (client_socket == -1) {
            perror("Failed to accept connection");
            continue; // Continue to next iteration instead of crashing
        }
        
        // Clear the buffer
        memset(buffer, 0, BUFFER_SIZE);
        
        // Read the HTTP request from the client
        ssize_t bytes_read = read(client_socket, buffer, BUFFER_SIZE - 1);
        if (bytes_read > 0) {
            buffer[bytes_read] = '\0'; // Null-terminate the string
            
            // Handle the request
            handle_request(client_socket, buffer);
        }
        
        // Close the client connection
        close(client_socket);
    }
    
    // Clean up (this code is unreachable in the current implementation)
    close(server_socket);
    return 0;
}
