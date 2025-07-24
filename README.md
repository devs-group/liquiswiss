# LiquiSwiss

Liquidity Planning

- [Landing Page](https://liquiswiss.ch/)
- [Discord](https://discord.gg/7ckBNzskYh)

## Configuration

> We are planning to add [Mailpit](https://mailpit.axllent.org/) for local e-mail testing and replace fixer.io with a
> local mock

1. Copy the [.env.example](.env.example)  into the same directory and name it `.env`
    - Set your credentials for [MariaDB](https://hub.docker.com/_/mariadb)
2. Copy the [backend/.env.example](backend/.env.example) into the same directory and name it `.env`
    - The `WEB_HOST` is set to http://localhost:3000 and works for local dev
    - THE `JWT_KEY` can be anything for local development but make sure to set it for [production](#production)
    - Set your credentials for [MariaDB](https://hub.docker.com/_/mariadb) and make sure they
      match [.env.example](.env.example)
    - Set your Token for [SendGrid](https://app.sendgrid.com/)
        - This requires you to have an account with SendGrid
            - Create an API key [here](https://app.sendgrid.com/settings/api_keys)
        - You also need to have a [Dynamic Template](https://mc.sendgrid.com/dynamic-templates) and get the ID
            - The template for LiquiSwiss looks like this:
            - ![sendgrid.png](.readme/sendgrid.png "Dynamic Template")
            - The values to submit can be found in [this model](backend/pkg/models/mail.go)
            - You can find the usage in the [SendgridService](backend/internal/service/sendgrid_service.go)
    - Set your credentials for [Fixer](https://fixer.io/)
        - This requires you to have an account with Fixer
        - Copy the API key that you find in the [Dashboard](https://fixer.io/dashboard)
3. Copy the [fronted/.env.example](frontend/.env.example) into the same directory and name it `.env`
    - The `NUXT_API_HOST` works fine for local dev but must be changed for production

## Admin

We are using [phpMyAdmin](https://www.phpmyadmin.net/) with Docker to provide an interface to the database.

- Check out the [.env.example](.env.example) to see the values you can set (all starting with `PMA_`). You can find more
  information [here](https://hub.docker.com/_/phpmyadmin)
- You can check out your database (locally) at: http://localhost:8082/

# Frontend (Nuxt 3)

Look at the [Nuxt 3 documentation](https://nuxt.com/docs/getting-started/introduction) to learn more.

> Make sure you are in the [frontend](frontend) directory for all the following actions

## Setup

```
npm install
```

## Development Server

```
npm run dev or npm run dev-host (to expose host and be able to connect from another device)
```

# Backend (Golang)

> Make sure you are in the [backend](backend) directory for all the following actions

## Setup

```
go get OR go mod tidy
```

## Development Server

1. Install [Air](https://github.com/air-verse/air): `go install github.com/air-verse/air@latest`

```
air
```

## Migrations

1. Install [Goose](https://github.com/pressly/goose): `go install github.com/pressly/goose/v3/cmd/goose@latest`

We differentiate between static and dynamic migrations whereas **static migrations** are all migrations that
actually hold data later and a migration down would lead to data loss such as table creation or alterations.

Dynamic migrations are stored functions, views or triggers, basically things that can be removed entirely and reapplied.

> The placeholder <directory> must either be replaced by "static" or "dynamic" (without quotes)

- Create Migration: `goose --dir internal/db/migrations/<directory> create <name-of-migration> sql`
    - Follow up with: `goose --dir internal/db/migrations/<directory> fix` to apply sequential numbering

> We run auto migrations on each app start. Since "air" will restart the app on any changes also in the .sql files
> the migrations will apply automatically but it might be helpful sometimes to rollback and reapply.

- Apply Migration: `goose --dir internal/db/migrations/<directory> mysql liquiswiss:password@/liquiswiss up`
- Rollback Migration: `goose --dir internal/db/migrations/<directory> mysql liquiswiss:password@/liquiswiss down`
- Or check out the [Makefile](backend/Makefile)

## Fixtures

> Optional step

You can fixtures from the [fixtures](backend/internal/service/db_service/fixtures) directory if you desire.
The dynamic migrations insert a minimal set of data required to make the app work properly. You can check out the
minimal inserted data in [00007_apply_minimal_fixtures.sql](backend/internal/db/migrations/dynamic/00007_apply_minimal_fixtures.sql)

## Tests

> Make sure you are in the [backend](backend) directory

> Make sure you spin up the test database with `docker compose up`

1. Install [Mockgen](https://github.com/uber-go/mock) with `go install go.uber.org/mock/mockgen@latest` to generate
   mocks
    - There are `go generate` commands already in the files so you can simply do `go generate ./...`
2. You can run all tests with `go test ./...` locally
3. Locally the [.env.local.testing](backend/.env.local.testing) is used
4. For the Github Action the [.env.github.testing](backend/.env.github.testing) is used
    - Check out the [ci.yml](.github/workflows/ci.yml) and check for the service used in the **test-backend** job
    - The environment variable `TESTING_ENVIRONMENT` determines which .env file to use

# Production

For production make sure you define the proper values for your envs (no matter in which way you provide them)

- `WEB_HOST` - Reflects your Frontend URL (eg. https://yourdomain.com)
- `JWT_KEY` - Should be a long and secure password
- `NUXT_API_HOST` - Reflects your Backend URL (eg https://api.yourdomain.com)