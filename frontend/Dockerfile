FROM public.ecr.aws/docker/library/node:16-alpine AS builder

ARG TARGETARCH
ARG TARGETOS=linux

WORKDIR /app
RUN apk add --no-cache git

COPY package*.json ./
RUN npm install --legacy-peer-deps

COPY . .
ENV TSC_COMPILE_ON_ERROR=true
ENV DISABLE_ESLINT_PLUGIN=true
ARG REACT_APP_API_URL
ENV REACT_APP_API_URL=$REACT_APP_API_URL
RUN npm run build

FROM public.ecr.aws/docker/library/node:16-alpine

WORKDIR /app
RUN npm install -g serve@14

COPY --from=builder /app/build ./build

EXPOSE 3000

CMD ["serve", "-s", "build", "-l", "3000"]
