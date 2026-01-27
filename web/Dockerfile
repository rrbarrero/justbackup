FROM golang:1.25.4-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG APP_URL
RUN go build -ldflags "-X main.DefaultBackendURL=${APP_URL}" -o justbackup cmd/client/main.go

FROM node:22-alpine AS base

# Install dependencies only when needed
FROM base AS deps


WORKDIR /app

COPY web/package.json web/package-lock.json* ./
RUN npm ci

# Rebuild the source code only when needed
FROM base AS builder

ARG APP_URL
ARG NEXT_PUBLIC_WS_URL
ARG BACKEND_INTERNAL_URL
ENV APP_URL=${APP_URL}
ENV NEXT_PUBLIC_WS_URL=${NEXT_PUBLIC_WS_URL}
ENV BACKEND_INTERNAL_URL=${BACKEND_INTERNAL_URL}
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY web .

# Next.js collects completely anonymous telemetry data about general usage.
# Learn more here: https://nextjs.org/telemetry
# Uncomment the following line in case you want to disable telemetry during the build.
# ENV NEXT_TELEMETRY_DISABLED 1

RUN npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app


ENV NODE_ENV production

# Uncomment the following line in case you want to disable telemetry during runtime.
# ENV NEXT_TELEMETRY_DISABLED 1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=go-builder /app/justbackup ./public/downloads/justbackup

# Set the correct permission for prerender cache
RUN mkdir .next
RUN chown nextjs:nodejs .next

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000
# set hostname to localhost
ENV HOSTNAME "0.0.0.0"

CMD ["node", "server.js"]
