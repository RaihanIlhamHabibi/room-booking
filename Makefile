.PHONY: up watch down reset-db setup test typecheck build health

up:
	docker compose up --build -d

watch:
	docker compose up --build --watch

down:
	docker compose down

setup:
	docker compose run --build --rm setup

# Deliberately removes only this standalone Compose project's database volume.
reset-db:
	docker compose down --volumes
	docker compose up --build -d

test:
	docker compose run --build --rm --no-deps api go test ./...
	docker compose run --build --rm --no-deps web corepack pnpm test

typecheck:
	docker compose run --build --rm --no-deps web corepack pnpm typecheck

build:
	docker compose run --build --rm --no-deps web corepack pnpm build

health:
	docker compose exec -T web node -e "fetch('http://kong:8000/api/health').then(r => { if (!r.ok) process.exit(1); return r.text(); }).then(console.log).catch(() => process.exit(1))"
