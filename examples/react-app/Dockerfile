# Build stage - builds from project root to access SDK
FROM node:20-alpine AS builder

# Build arguments for Vite environment variables
ARG VITE_API_URL=http://localhost:8001
ARG VITE_GOOGLE_CLIENT_ID

# Set environment variables for build
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID

WORKDIR /app

# Copy package files for SDK
COPY packages/react/package.json packages/react/
COPY packages/react/tsconfig.json packages/react/
COPY packages/react/src packages/react/src/

# Copy package files for example app
COPY examples/react-app/package.json examples/react-app/
COPY examples/react-app/tsconfig.json examples/react-app/
COPY examples/react-app/tsconfig.node.json examples/react-app/
COPY examples/react-app/vite.config.ts examples/react-app/
COPY examples/react-app/index.html examples/react-app/

# Install SDK dependencies and build
WORKDIR /app/packages/react
RUN npm install
RUN npm run build

# Install example app dependencies
WORKDIR /app/examples/react-app
RUN npm install

# Copy example app source
COPY examples/react-app/src ./src/

# Build the app (Vite will embed VITE_* env vars)
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy custom nginx config
COPY examples/react-app/nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets from builder
COPY --from=builder /app/examples/react-app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
