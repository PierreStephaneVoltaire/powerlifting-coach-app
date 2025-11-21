FROM public.ecr.aws/docker/library/node:16-alpine AS builder

ARG TARGETARCH=amd64
ARG TARGETOS=linux

WORKDIR /app
RUN apk add --no-cache git

COPY package*.json ./
RUN npm install --legacy-peer-deps

COPY . .
ENV TSC_COMPILE_ON_ERROR=true
ENV DISABLE_ESLINT_PLUGIN=true
RUN npm run build

FROM public.ecr.aws/docker/library/node:16-alpine

WORKDIR /app
RUN npm install -g serve@14

COPY --from=builder /app/build ./build
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["/entrypoint.sh"]
CMD ["serve", "-s", "build", "-l", "3000"]
