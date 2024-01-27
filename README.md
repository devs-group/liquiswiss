# LiquiSwiss

Liquidity Planning

# Frontend (Nuxt 3)

Look at the [Nuxt 3 documentation](https://nuxt.com/docs/getting-started/introduction) to learn more.

## Setup

Make sure to install the dependencies:

```bash
cd fontend
npm install

or

docker compose build (if you have docker)
```

## Development Server

Start the development server:

```bash
cd frontend
npm run dev

or

cd frontend
npm run dev-host (to expose host and be able to connect from another device)

or

docker compose up -d (if you have docker)
```

# Backend (Strapi)

Look at the [Strapi documentation](https://docs.strapi.io/dev-docs/faq) to learn more.

## Setup

```bash
cd backend
npm install

or

docker compose build (if you have docker)
```

## Development Server

Start the development server to be able to create content types:

```bash
cd backend
npm run develop

or

docker compose up -d (if you have docker)
```

## Troubleshooting

If you have issues with generated code try to run this command:

```bash
npm run strapi ts:generate-types
```

## Production

Build the application for production:

Every push or merge to master will trigger an automatic deployment via Gitlab Pipeline
