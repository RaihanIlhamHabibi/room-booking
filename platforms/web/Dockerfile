FROM node:24-alpine

WORKDIR /app
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY platforms/web/package.json platforms/web/package.json
COPY packages/api/package.json packages/api/package.json
COPY packages/features/package.json packages/features/package.json
COPY packages/ui/package.json packages/ui/package.json
RUN corepack pnpm install --frozen-lockfile
COPY . .

EXPOSE 3000
CMD ["corepack", "pnpm", "--filter", "@room-booking/web", "dev", "--hostname", "0.0.0.0"]
