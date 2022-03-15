FROM node
RUN mkdir /app
WORKDIR /app
RUN npm install web3
ADD ./main.js main.js
ADD ./package.json package.json
RUN npm install -g .
