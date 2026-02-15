FROM node:20-alpine AS base
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .

FROM base AS local
EXPOSE 1313
CMD ["npm", "run", "dev"]

FROM base AS build
ARG VITE_API_URL
ARG VITE_APP_URL
ARG VITE_ENV
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_APP_URL=$VITE_APP_URL
ENV VITE_ENV=$VITE_ENV
RUN npm run build

FROM node:20-alpine AS prod
WORKDIR /app
RUN npm install -g serve@14
COPY --from=build /app/dist ./dist
EXPOSE 1313
CMD ["serve", "-s", "dist", "-l", "1313"]
