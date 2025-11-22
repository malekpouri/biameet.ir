FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY frontend/package.json ./
RUN npm install

# Copy source
COPY frontend .

# Build CSS
RUN npm run build

# Copy static assets to dist (simple build step)
RUN cp src/index.html dist/ && cp src/app.js dist/

# Final stage
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf
# Optional: Copy custom nginx config if needed
# COPY docker/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
