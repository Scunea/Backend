import express from "express";
import { Client } from 'pg';

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, Map<string, WebSocket>>>) => {
    app.get('/pendingchildren', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                res.send(dbRes.rows.filter(x => JSON.parse(x.pendingparents).includes(res.locals.user)).map(x => {
                    return {
                        id: x.id,
                        name: x.name
                    };
                }));
            } else {
                console.log(err);
                res.status(500).send({ error: "Server error." });
            }
        });
    });

    app.get('/pendingparents', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                res.send(JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).pendingparents).map((x: string) => {
                    const user = dbRes.rows.find(y => y.id === x);
                    return {
                        id: user.id,
                        name: user.name
                    };
                }));
            } else {
                console.log(err);
                res.status(500).send({ error: "Server error." });
            }
        });
    });

    app.put('/parents', async (req: express.Request, res: express.Response) => {
        if (req.body.email) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    const user = dbRes.rows.find(x => x.id === res.locals.user);
                    const parent = dbRes.rows.find(x => x.email === req.body.email);
                    if (user && parent) {
                        let pendingparents = JSON.parse(user.pendingparents);
                        pendingparents.push(parent.id);
                        user.pendingparents = JSON.stringify(pendingparents);

                        database.query(`UPDATE users SET pendingparents = $1 WHERE id = $2`, [user.pendingparents, user.id], (err, dbResp) => {
                            if (!err) {
                                Array.from(websockets.get(res.locals.school)?.get(user.id)?.values() ?? [])?.forEach(websocket => {
                                    websocket.send(JSON.stringify({
                                        event: 'parentInvited', parent: {
                                            id: parent.id,
                                            name: parent.name
                                        }
                                    }));
                                });
                                Array.from(websockets.get(res.locals.school)?.get(parent.id)?.values() ?? [])?.forEach(websocket => {
                                    websocket.send(JSON.stringify({
                                        event: 'childrenInvited', children: {
                                            id: user.id,
                                            name: user.name
                                        }
                                    }));
                                });
                                res.send({});
                            } else {
                                console.log(err);
                                res.status(500).send({ error: "Server error." });
                            }
                        });
                    } else {
                        res.status(404).send({ error: "Not found." });
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

    app.delete('/parents/*', async (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const parentId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        if (parentId) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    const user = dbRes.rows.find(x => x.id === res.locals.user);
                    if (user) {
                        let pendingparents = JSON.parse(user.pendingparents);
                        if (pendingparents.includes(parentId)) {
                            pendingparents.splice(pendingparents.indexOf(parentId), 1);
                        }
                        user.pendingparents = JSON.stringify(pendingparents);

                        database.query(`UPDATE users SET pendingparents = $1 WHERE id = $2`, [user.pendingparents, user.id], (err, dbResp) => {
                            if (!err) {
                                Array.from(websockets.get(res.locals.school)?.get(user.id)?.values() ?? [])?.forEach(websocket => {
                                    websocket.send(JSON.stringify({ event: 'parentInviteRemoved', id: parentId }));
                                });
                                Array.from(websockets.get(res.locals.school)?.get(parentId)?.values() ?? [])?.forEach(websocket => {
                                    websocket.send(JSON.stringify({ event: 'childrenInviteRemoved', id: user.id }));
                                });
                                res.send({});
                            } else {
                                console.log(err);
                                res.status(500).send({ error: "Server error." });
                            }
                        });
                    } else {
                        res.status(404).send({ error: "Not found." });
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

    app.post('/accept/*', async (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const childId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        if (childId) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    const user = dbRes.rows.find(x => x.id === childId);

                    let pendingparents = JSON.parse(user.pendingparents);
                    if (pendingparents.includes(res.locals.user)) {
                        pendingparents.splice(pendingparents.indexOf(res.locals.user), 1);
                        user.pendingparents = JSON.stringify(pendingparents);

                        let parents = JSON.parse(user.parents);
                        parents.push(res.locals.user);
                        user.parents = JSON.stringify(parents);

                        database.query(`UPDATE users SET pendingparents = $1, parents = $2 WHERE id = $3`, [user.pendingparents, user.parents, user.id], (err, dbResp) => {
                            if (!err) {
                                Array.from(websockets.get(res.locals.school)?.get(childId)?.values() ?? [])?.forEach(websocket => {
                                    websocket.send(JSON.stringify({ event: 'parentInviteRemoved', id: res.locals.user }));
                                });
                                Array.from(websockets.get(res.locals.school)?.get(res.locals.user)?.values() ?? [])?.forEach(websocket => {
                                    websocket.send(JSON.stringify({ event: 'childrenInviteRemoved', id: childId }));
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