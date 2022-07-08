import express from "express";
import { Client } from 'pg';
import crypto from 'crypto';

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, Map<string, WebSocket>>>) => {
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

                    let pendingschools = JSON.parse(user.pendingschools);
                    if (pendingschools.includes(schoolId)) {
                        pendingschools.splice(pendingschools.indexOf(schoolId), 1);
                        user.pendingschools = JSON.stringify(pendingschools);

                        let schools = JSON.parse(user.schools);
                        schools.push(schoolId);
                        user.schools = JSON.stringify(schools);

                        database.query(`UPDATE users SET pendingschools = $1, schools = $2 WHERE id = $3`, [user.pendingschools, user.schools, user.id], (err, dbResp) => {
                            if (!err) {

                                dbRes.rows.forEach((userToSend: any) => {
                                    if (JSON.parse(user.teacher)[schoolId] || JSON.parse(userToSend.administrator).includes(schoolId)) {
                                        Array.from(websockets.get(schoolId)?.get(userToSend.id)?.values() ?? [])?.forEach(websocket => {
                                            websocket.send(JSON.stringify({ event: 'deletedUser', userId: user.id }));
                                            websocket.send(JSON.stringify({
                                                event: 'newUser', user: {
                                                    id: user.id,
                                                    name: user.name,
                                                    email: user.email,
                                                    subject: JSON.parse(user.teacher)[schoolId],
                                                    children: dbRes.rows.filter(x => JSON.parse(x.parents).includes(user.id))?.map(x => x.id),
                                                    type: JSON.parse(dbRes.rows.find(y => y.id === user.id).administrator).includes(schoolId) ?
                                                    'Administrator' : JSON.parse(dbRes.rows.find(y => y.id === user.id).teacher)[schoolId] ?
                                                        'Teacher' : dbRes.rows.find(y => JSON.parse(y.schools).includes(schoolId) &&  JSON.parse(y.parents)?.includes(user.id)) ?
                                                            'Parent' :
                                                            'Student',
                                                }
                                            }));
                                        });
                                    }
                                });

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