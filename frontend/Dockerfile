# syntax=docker/dockerfile:1.7

# ── Stage 1: build — React production bundle
FROM node:20-alpine AS builder

WORKDIR /app

# Copia prima i manifest: consente cache layer su `yarn install` finché non
# cambiano le deps.
COPY package.json yarn.lock ./

# Yarn classic preinstallato nell'immagine node:20-alpine. `--frozen-lockfile`
# fallisce se il lockfile è disallineato (previene build non riproducibili).
RUN yarn install --frozen-lockfile --network-timeout 600000

# API URL configurabile al build time (CRA: REACT_APP_*).
# In dev lasciamo vuoto (SPA usa URL relativo proxy-ato dal nginx a /api).
# Vite build-time env (era REACT_APP_BACKEND_URL su CRA, ora VITE_BACKEND_URL).
# Default vuoto = SPA usa URL relativi (`/api/...`), proxati da nginx in prod.
ARG VITE_BACKEND_URL=""
ENV VITE_BACKEND_URL=$VITE_BACKEND_URL

COPY . /app

RUN yarn build


# ── Stage 2: nginx runtime — serve SPA + reverse proxy verso backend
FROM nginx:1.27-alpine AS runtime

# Config nginx copiata dal build context (siblings di Dockerfile)
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx/default.conf /etc/nginx/conf.d/default.conf

# Bundle statico
COPY --from=builder /app/build /usr/share/nginx/html

EXPOSE 80 443

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost/healthz || exit 1

# nginx nel foreground (default dell'immagine, esplicitato per chiarezza)
CMD ["nginx", "-g", "daemon off;"]
