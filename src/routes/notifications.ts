import express from 'express';
import { Client } from 'pg';

export default (app: express.Application, database: Client) => {
    app.post('/notifications', async (req: express.Request, res: express.Response) => {
        if (req.body.endpoint && req.body.keys?.p256dh && req.body.keys?.auth) {
            database.query(`INSERT INTO notifications (endpoint, p256dh, auth, id) VALUES($1, $2, $3, $4) ON CONFLICT DO NOTHING`, [req.body.endpoint, req.body.keys.p256dh, req.body.keys.auth, res.locals.user], (err, dbRes) => {
                if (!err) {
                    res.send({});
                } else {
                    console.log(err);
                    res.status(500).send({ error: "Server error." });
                }
            });
        } else {
            res.status(400).send({});
        }
    });
};