# Build stage
FROM node:22-alpine AS builder

WORKDIR /app

# node:22-alpine ships npm 10.9.8, which mis-resolves this project's
# optional-peer-dependency graph (@bomb.sh/tab's optional commander peer
# collides with an unrelated nested commander@7.x) and makes `npm ci` fail
# with EUSAGE ("lock file's commander@7.2.0 does not satisfy
# commander@15.0.0") even though the lockfile itself is valid — confirmed
# fixed under npm 11+.
RUN npm install -g npm@11

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
