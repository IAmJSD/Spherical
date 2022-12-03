FROM spherical_build
FROM mcr.microsoft.com/playwright:v1.27.0-focal
WORKDIR /app
COPY package.json /app/package.json
COPY package-lock.json /app/package-lock.json
RUN npm ci
RUN npm run browsers:get
COPY . /app
COPY --from=0 /app/bin /app/bin
ENTRYPOINT ["npm", "start"]
