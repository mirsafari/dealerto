//Initialize config, express and winston for logging
const appConfig = require('./config/server-config.js')

const winston = require('winston');
const expressWinston = require('express-winston');

const express = require('express');

//Init express and winston
const app = express();

const logger = winston.createLogger({
    transports: [
        new winston.transports.Console()
    ]
});

//Configure and start winston express-middleware logging
app.use(expressWinston.logger({
    transports: [
      new winston.transports.Console()
    ],
    format: winston.format.combine(
      winston.format.json()
    ),
    meta: appConfig.metaLogging,
    msg: "HTTP {{res.statusCode}} {{req.method}} {{res.responseTime}}ms {{req.url}}", 
    expressFormat: true, 
    colorize: false 
  }));

//Add body parser middleware
app.use(express.json());

app.use('/api/alerts', require('./routes/api/alerts'));


// Start HTTP server
app.listen(appConfig.httpPort, () => logger.info(`Example app listening at http://localhost:${appConfig.httpPort}`))