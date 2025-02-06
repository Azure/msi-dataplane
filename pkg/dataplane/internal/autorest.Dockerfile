FROM node:lts

RUN npm install -g autorest

ENTRYPOINT ["autorest"]
