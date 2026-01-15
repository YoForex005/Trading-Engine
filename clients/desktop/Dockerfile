# Multi-stage Dockerfile for React + Vite frontend
# Build stage compiles React app, runtime serves static files with nginx

# Build stage - compile React application
FROM node:20-alpine AS builder

WORKDIR /app

# Copy dependency files first for layer caching
# Dependencies only rebuild when package.json or bun.lock change
COPY package.json bun.lock ./

# Install bun and dependencies
RUN npm install -g bun && bun install --frozen-lockfile

# Copy all source code
COPY . .

# Build for production
# Vite outputs optimized bundle to dist/
ENV NODE_ENV=production
RUN bun run build

# Runtime stage - serve with nginx
FROM nginx:1.25-alpine

# Remove default nginx configuration
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx configuration for SPA routing
COPY nginx.conf /etc/nginx/conf.d/

# Copy built static files from build stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Frontend port
EXPOSE 80

# Start nginx in foreground
CMD ["nginx", "-g", "daemon off;"]
