# Build stage
FROM node:20-alpine AS builder

# Build arguments for Vite environment variables
ARG VITE_API_URL=http://localhost:8001
ARG VITE_CASDOOR_ENDPOINT=http://localhost:8000
ARG VITE_CASDOOR_CLIENT_ID
ARG VITE_CASDOOR_ORGANIZATION=built-in
ARG VITE_CASDOOR_APPLICATION=app-built-in

# Set environment variables for build
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_CASDOOR_ENDPOINT=$VITE_CASDOOR_ENDPOINT
ENV VITE_CASDOOR_CLIENT_ID=$VITE_CASDOOR_CLIENT_ID
ENV VITE_CASDOOR_ORGANIZATION=$VITE_CASDOOR_ORGANIZATION
ENV VITE_CASDOOR_APPLICATION=$VITE_CASDOOR_APPLICATION

WORKDIR /app

# Copy package files
COPY package.json ./

# Install dependencies
RUN npm install

# Copy source files
COPY . .

# Build the app
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
