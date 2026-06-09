# Go Watermarking

A web application and API for adding custom watermarks to images. Built with a Go backend and a SvelteKit frontend, this application allows you to process multiple images in batches and customize the watermark's text, image, color, scale, angle, and opacity.

## Features

- **Text & Image Watermarks**: Choose between overlaying custom text or an existing image as a watermark.
- **Batch Processing**: Upload and process multiple images simultaneously.
- **Customization**: Adjust scale, rotation angle, opacity, color, and placement mode.
- **Single Binary**: The SvelteKit frontend is compiled and embedded directly into the Go binary for easy deployment.

## Tech Stack

- **Backend**: Go (Golang) 1.26
- **Frontend**: Svelte 5, SvelteKit (Static Adapter), Vite
- **Styling**: TailwindCSS 4, DaisyUI
- **Package Manager**: Bun (or npm)

## Getting Started

### Prerequisites

- Go 1.26+
- Bun (or Node.js/npm)

### Development

To run the application in development mode, you will need to start both the backend server and the frontend development server.

1. **Start the Backend Server**

   ```bash
   cd server
   go run cmd/server/main.go
   ```
   *The backend will run on `http://localhost:8080`.*

2. **Start the Frontend Development Server**

   Open a new terminal window:
   ```bash
   cd web
   bun install
   bun run dev
   ```
   *The frontend will run on `http://localhost:5173` and connect to the backend API.*

### Production Build

For production, the frontend is built into a static bundle and embedded into the Go binary. This gives you a single executable to run the entire application.

1. **Build the Frontend**

   ```bash
   cd web
   bun install
   bun run build
   ```
   *This compiles the SvelteKit app and outputs the static files into `server/internal/web/dist`.*

2. **Build the Go Binary**

   ```bash
   cd ../server
   go build -o go-watermarking cmd/server/main.go
   ```

3. **Run the Application**

   ```bash
   ./go-watermarking
   ```
   *The server will start and serve both the frontend UI and the API on `http://localhost:8080`.*

## Project Structure

- `/server` - Go backend API and image processing logic.
- `/web` - SvelteKit frontend application.
