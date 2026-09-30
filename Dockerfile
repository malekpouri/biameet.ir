# Single image: the Go binary serves the API, the pages and the pre-compressed
# assets itself, so no nginx container is needed.

# 1. Frontend: Tailwind CSS, bundled JS and the font -> backend/web/dist
FROM node:22-alpine AS web
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
RUN OUT_DIR=/src/backend/web/dist npm run build

# 2. Backend: one static binary with the frontend embedded
FROM golang:1.24-alpine AS app
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=web /src/backend/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /biameet ./cmd \
    && /biameet version \
    && mkdir /data

# 3. Runtime: nothing but the binary. It starts as root only to take ownership
#    of /data (volumes from 1.x are root-owned), then switches to RUN_AS_UID
#    before opening the database or serving anything.
FROM scratch
COPY --from=app /biameet /biameet
COPY --from=app --chown=65534:65534 /data /data
ENV PORT=8080 \
    DB_PATH=/data/biameet.db \
    RUN_AS_UID=65534 \
    GOMEMLIMIT=48MiB
EXPOSE 8080
VOLUME /data
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s CMD ["/biameet", "healthcheck"]
ENTRYPOINT ["/biameet"]
