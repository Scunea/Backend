import { Client } from 'pg';
import { createServer } from 'http';
import { WebSocketServer } from 'ws';

import app from './app';
import config from './utils/config';
import { createLogger } from './utils/logger';

import keygen from './utils/keygen';
import vapid from './utils/vapid';
import ws from './utils/ws';
import routes from './routes';
import db from './utils/db';

const wss = new WebSocketServer({ noServer: true });
const database = new Client({
  user: config.db.user,
  host: config.db.host,
  database: config.db.database,
  password: config.db.password,
  port: config.db.port,
});

const logger = createLogger(config.env === 'development');
const server = createServer(app);

const websockets = new Map();

keygen(logger);

vapid(config.vapid);

ws(wss, websockets, server, database);

routes(app, database, websockets);

server.listen(config.server.port, async () => {
  db(database, logger);

  logger.info(`Listening on port ${config.server.port}`);
});