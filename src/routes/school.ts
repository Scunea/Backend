import express from "express";
import { Client } from 'pg';
import crypto from 'crypto';

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, WebSocket[]>>) => {
    app.post('/create', async (req: express.Request, res: express.Response) => {
        if (req.body.name) {
            const uuid = crypto.randomUUID();
            database.query(`INSERT INTO schools (id, name) VALUES($1, $2)`, [uuid, req.body.name], (err, dbRes) => {
                if (!err) {
                    database.query(`SELECT * FROM users`, async (err, dbRes) => {
                        if (!err) {
                            let user = dbRes.rows.find(x => x.id === res.locals.user);

                            let schools = JSON.parse(user.schools);
                            schools.push(uuid);
                            user.schools = JSON.stringify(schools);

                            let administrator = JSON.parse(user.administrator);
                            administrator.push(uuid);
                            user.administrator = JSON.stringify(administrator);
                            database.query(`UPDATE users SET schools = $1, administrator = $2 WHERE id = $3`, [user.schools, user.administrator, user.id], (err, dbResp) => {
                                if (!err) {
                                    res.send({
                                        id: uuid,
                                        name: req.body.name
                                    });
                                } else {
                                    console.log(err);
                                    res.status(500).send({ error: "Server error." });
                                }
                            });
                        } else {
                            console.log(err);
                            res.status(500).send({ error: "Server error." });
                        }
                    });
                } else {
                    console.log(err);
                    res.status(500).send({ error: "Server error." });
                }
            });
        } else {
            res.status(400).send({ error: "Missing required argument." });
        }
    });

    app.post('/join/*', async (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const schoolId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        if (schoolId) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    const user = dbRes.rows.find(x => x.id === res.locals.user);

                    let pending = JSON.parse(user.pending);
                    if (pending.includes(schoolId)) {
                        pending.splice(pending.indexOf(schoolId), 1);
                        user.pending = JSON.stringify(pending);

                        let schools = JSON.parse(user.schools);
                        schools.push(schoolId);
                        user.schools = JSON.stringify(schools);

                        database.query(`UPDATE users SET pending = $1, schools = $2 WHERE id = $3`, [user.pending, user.schools, user.id], (err, dbResp) => {
                            if (!err) {
                                res.send({});
                            } else {
                                console.log(err);
                                res.status(500).send({ error: "Server error." });
                            }
                        });
                    } else {
                        res.status(401).send({ error: "Not authorized." });
                    }
                } else {
                    console.log(err);
                    res.status(500).send({ error: "Server error." });
                }
            });
        } else {
            res.status(400).send({ error: "Missing required argument." });
        }
    });
};