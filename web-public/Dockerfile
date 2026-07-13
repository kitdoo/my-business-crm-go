# Build stage
FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY ./ ./
RUN npm run build

# Final stage
FROM node:22-alpine

ENV NODE_ENV=production
ENV NITRO_PORT=3000
ENV NITRO_HOST=0.0.0.0

WORKDIR /app
COPY --from=builder /app/.output ./.output
COPY --from=builder /app/proto ./proto

EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
