# syntax=docker/dockerfile:1.7
# ─── Build stage ─────────────────────────────────────────────────────────────
# RouteFlow ships as a Tauri desktop application; this Docker image bundles
# only the Vite web build for the browser preview / staging deployment.
# To build a native desktop binary, run `npm run tauri:build` on the host.
FROM node:22-alpine AS builder

WORKDIR /src

COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund

COPY tsconfig.json vite.config.ts tailwind.config.ts postcss.config.js \
     index.html eslint.config.js .prettierrc.json ./
COPY src ./src

ARG VITE_API_BASE_URL=
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

RUN npm run build

# ─── Runtime stage ───────────────────────────────────────────────────────────
FROM nginx:1.27-alpine AS runtime

COPY --from=builder /src/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

# Healthcheck: nginx must serve the SPA shell.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://127.0.0.1/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
