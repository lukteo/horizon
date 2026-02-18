# Use Alpine base image with Bun
FROM oven/bun:latest

# Set working directory
WORKDIR /app

# Copy package files
COPY client/package*.json ./

# Install dependencies
RUN bun install --production

# Copy source code
COPY client/ .

# Build the application
RUN bun run build

# Expose port
EXPOSE 3000

# Start the application
CMD ["bun", "start"]