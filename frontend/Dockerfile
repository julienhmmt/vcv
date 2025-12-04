FROM node:25.1-alpine

WORKDIR /app

RUN apk add --no-cache git && rm -rf /var/cache/apk/*

COPY package*.json ./
RUN npm ci --prefer-offline --no-audit --no-fund

COPY . .

EXPOSE 4321

CMD ["sh", "-c", "npx astro telemetry disable && npm run dev"]