# Go image
FROM golang:1.20.5

# Working directory
WORKDIR /todo-app

# Copy all application source code
COPY . .

# Download dependencies
RUN go mod download

# Build the app
RUN make build

# Expose our port
EXPOSE 3000

# Run Executable
CMD [ "./bin/todo" ]