# Use an official Go image as the base
FROM golang:1.23.5

# Install Air for hot reloading
RUN go install github.com/air-verse/air@latest

# Set the working directory
WORKDIR /app

# Copy the Air configuration file
COPY .air.toml ./

# Copy the project files
COPY . .

# Install dependencies
RUN go mod download

# Expose port 8080
EXPOSE 8080

# Use Air for hot reloading
CMD ["air"]
