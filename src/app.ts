import compression from 'compression';
import cors from 'cors';
import express from 'express';
import helmet from 'helmet';
import rateLimit from 'express-rate-limit';
import config from './utils/config';

const app = express();

app.use(helmet());

app.use((req, res, next) => {
    if (req.path === "/upload") {
        next();
    }
    else {
        express.json()(req, res, next);
    }
});

app.use(compression());

app.use(cors());

if (config.env !== 'development') {
    app.use(rateLimit({
        windowMs: 60000,
        max: 100
    }));
}

export default app;