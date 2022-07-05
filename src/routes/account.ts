import express from "express";
import { Client } from 'pg';
import argon2 from 'argon2';

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, WebSocket[]>>) => {
    app.patch('/account', async (req: express.Request, res: express.Response) => {
        if ((req.body.name || req.body.password)) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    const user = dbRes.rows.find(x => x.id === res.locals.user);
                    user.name = req.body.name;
                    if (req.body.password) {
                        user.password = await argon2.hash(req.body.password, { type: argon2.argon2id });
                    }
                    if (await argon2.verify(user.password, req.body.currentPassword ?? '', { type: argon2.argon2id })) {
                        if (!user.administrator.includes(res.locals.school) && req.body.name) {
                            res.status(403).send({ error: "Not authorized." });
                            return;
                        }
                        database.query(`UPDATE users SET name = $1, password = $2 WHERE id = $3`, [user.name, user.password, user.id], (err, dbResp) => {
                            if (!err) {
                                dbRes.rows.forEach((userToSend: any) => {
                                    if (JSON.parse(userToSend.administrator).includes(res.locals.school) || res.locals.user === user.id) {
                                        const child = dbRes.rows.find(x => JSON.parse(x.parent)[res.locals.school]?.includes(user.id));
                                        let fixedChild: { id: string; name: string; };
                                        if (child) {
                                            fixedChild = {
                                                id: child.id,
                                                name: child.name
                                            }
                                        }
                                        websockets?.forEach(websocket => {
                                            if (websocket.has(userToSend.id)) {
                                                websocket.get(userToSend.id)?.forEach(websocket => {
                                                    websocket.send(JSON.stringify({
                                                        event: 'editedUser', user: {
                                                            id: user.id,
                                                            name: user.name,
                                                            subject: JSON.parse(user.teacher)[res.locals.school],
                                                            child: fixedChild,
                                                            type: JSON.parse(user.administrator).includes(res.locals.school) ?
                                                                'Administrator' : JSON.parse(user.teacher)[res.locals.school] ?
                                                                    'Teacher' : dbRes.rows.find(x => JSON.parse(x.parent)[res.locals.school]?.includes(user.id)) ?
                                                                        'Parent' :
                                                                        'Student'
                                                        }
                                                    }));
                                                });
                                            }
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
                        res.status(401).send({ error: "Invalid credentials." });
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