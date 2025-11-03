FROM public.ecr.aws/docker/library/node:18-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git 

COPY package*.json ./
RUN npm install --legacy-peer-deps

COPY . .
ENV TSC_COMPILE_ON_ERROR=true
ENV DISABLE_ESLINT_PLUGIN=true
RUN npm run build

FROM public.ecr.aws/docker/library/nginx:alpine

COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]