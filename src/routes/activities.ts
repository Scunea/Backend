import { File } from '../interfaces';

import express from "express";
import { Client } from 'pg';
import crypto from 'crypto';
import fs from 'fs';
import webpush from 'web-push';

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, Map<string, WebSocket>>>) => {
    app.get('/activities', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM activities`, async (err, dbRes) => {
            if (!err) {
                let activities = dbRes.rows.filter(x => x.school === res.locals.school);
                database.query(`SELECT * FROM users`, async (err, dbRes) => {
                    if (!err) {
                        if (JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).teacher)[res.locals.school]) {
                            activities = activities.filter((x: any) => x.author === res.locals.user);
                        } else if (!JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                            activities = activities.filter((x: any) => x.receiver.includes(res.locals.user));
                        }
                        res.send(activities.map(x => {
                            x.files = JSON.parse(x.files);
                            x.receiver = JSON.parse(x.receiver);
                            if (JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).teacher)[res.locals.school]) {
                                x.delivered = Object.keys(JSON.parse(x.delivered)).reduce((obj, key, index) => ({
                                    ...obj, [key]: Object.values(JSON.parse(x.delivered)).map((x: any) => {
                                        x.date = Number(x.date);
                                        x.name = dbRes.rows.find(y => y?.id === key)?.name ?? "Deleted user";
                                        return x;
                                    })[index]
                                }), {});
                                x.viewed = JSON.parse(x.viewed);
                                x.result = JSON.parse(x.result);
                            } else {
                                x.delivered = JSON.parse(x.delivered)[res.locals.user];
                                x.viewed = JSON.parse(x.viewed)[res.locals.user];
                                x.result = JSON.parse(x.result)[res.locals.user];
                            }
                            x.subject = JSON.parse(dbRes.rows.find(y => y?.id === x.author).teacher)[res.locals.school];
                            x.author = { id: x.author, name: dbRes.rows.find(y => y?.id === x.author)?.name ?? "Deleted user" };
                            x.date = Number(x.date);
                            x.expiration = Number(x.expiration);
                            delete x.school;
                            return x;
                        }));
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
    });

    app.post('/activities', (req: express.Request, res: express.Response) => {
        const files = fs.readdirSync(__dirname + '/../../files');
        const activity = {
            id: crypto.randomUUID(),
            title: req.body.title,
            description: req.body.description ?? '',
            files: req.body.files ?? [],
            type: req.body.type ?? '',
            delivery: req.body.delivery ?? '',
            author: res.locals.user,
            date: Date.now(),
            expiration: !isNaN(req.body.expiration) ? req.body.expiration : 0,
            receiver: req.body.receiver ?? [],
            delivered: {},
            result: {},
            viewed: {},
            school: res.locals.school
        };
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                let avaliableUsers = dbRes.rows.map(x => x.id);
                if (JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).teacher)[res.locals.school]) {
                    if (!activity.receiver.some((x: string) => !avaliableUsers.includes(x))) {
                        const users = dbRes.rows;
                        if (activity.title && activity.receiver.length > 0 && !activity.files.map((x: File) => !!files.find((y: string) => y.startsWith(x.id))).includes(false)) {
                            database.query(`INSERT INTO activities (id, title, description, files, type, delivery, author, date, expiration, receiver, school, delivered, result, viewed) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`, [activity.id, activity.title, activity.description, JSON.stringify(activity.files), activity.type, activity.delivery, activity.author, activity.date.toString(), activity.expiration.toString(), JSON.stringify(activity.receiver), activity.school, JSON.stringify(activity.delivered), JSON.stringify(activity.result), activity.viewed], (err, dbRes) => {
                                if (!err) {
                                    let websocketiedActivity: any = { ...activity };
                                    websocketiedActivity.subject = JSON.parse(users.find(y => y?.id === websocketiedActivity.author).teacher)[res.locals.school];
                                    websocketiedActivity.author = { id: websocketiedActivity.author, name: users.find(y => y?.id === websocketiedActivity.author)?.name ?? "Deleted user" };

                                    const receivers = users.filter(x => JSON.parse(x.administrator).includes(res.locals.school)).map(x => x.id).concat([...activity.receiver, activity.author]);
                                    receivers.forEach((receiver: string) => {
                                        Array.from(websockets.get(res.locals.school)?.get(receiver)?.values() ?? [])?.forEach(websocket => {
                                            websocket.send(JSON.stringify({ event: 'newActivity', ...websocketiedActivity }));
                                        });
                                    });

                                    database.query(`SELECT * FROM notifications`, async (err, dbRes) => {
                                        if (!err) {
                                            dbRes.rows.filter(x => receivers.includes(x.id)).forEach(row => {
                                                webpush.sendNotification({
                                                    endpoint: row.endpoint,
                                                    keys: {
                                                        p256dh: row.p256dh,
                                                        auth: row.auth
                                                    }
                                                }, JSON.stringify({ event: 'newActivity', ...websocketiedActivity })).catch(err => {
                                                    if (err.statusCode === 404 || err.statusCode === 410) {
                                                        database.query(`DELETE FROM notifications WHERE endpoint = $1`, [row.endpoint], async (err, dbRes) => { });
                                                    }
                                                });
                                            });
                                        }
                                    });
                                    res.status(201).send({});
                                } else {
                                    console.log(err);
                                    res.status(500).send({ error: "Server error." });
                                }
                            });
                        } else {
                            res.status(400).send({ error: "Missing required argument." });
                        }
                    } else {
                        res.status(403).send({ error: "Not authorized." });
                    }
                } else {
                    res.status(403).send({ error: "Not authorized." });
                }
            } else {
                console.log(err);
                res.status(500).send({ error: "Server error." });
            }
        });
    });

    app.patch('/activities/*', (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const activityId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        const files = fs.readdirSync(__dirname + '/../../files');
        database.query(`SELECT * FROM activities`, async (err, dbRes) => {
            if (!err) {
                const oldActivity = dbRes.rows.find(x => x.id === activityId);
                const newActivity = {
                    title: req.body.title,
                    description: req.body.description ?? '',
                    files: req.body.files ?? [],
                    type: req.body.type ?? '',
                    delivery: req.body.delivery ?? '',
                    expiration: !isNaN(req.body.expiration) ? req.body.expiration : 0,
                    receiver: req.body.receiver ?? [],
                };
                database.query(`SELECT * FROM users`, async (err, dbResu) => {
                    if (!err) {
                        if (oldActivity.author === res.locals.user || JSON.parse(dbResu.rows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                            let avaliableUsers = dbResu.rows.map(x => x.id);
                            if (JSON.parse(dbResu.rows.find(x => x.id === res.locals.user).teacher)[res.locals.school]) {
                                if (!newActivity.receiver.some((x: string) => !avaliableUsers.includes(x))) {
                                    if (newActivity.title && newActivity.receiver.length > 0 && !newActivity.files.map((x: File) => !!files.find((y: string) => y.startsWith(x.id))).includes(false)) {
                                        database.query(`UPDATE activities SET title = $1, description = $2, files = $3, type = $4, delivery = $5, expiration = $6, receiver = $7 WHERE id = $8`, [newActivity.title, newActivity.description, JSON.stringify(newActivity.files), newActivity.type, newActivity.delivery, newActivity.expiration.toString(), JSON.stringify(newActivity.receiver), activityId], (err, dbRes) => {
                                            if (!err) {
                                                dbResu.rows.filter(x => JSON.parse(x.administrator).includes(res.locals.school)).map(x => x.id).concat([...newActivity.receiver, oldActivity.author]).forEach((receiver: string) => {
                                                    Array.from(websockets.get(res.locals.school)?.get(receiver)?.values() ?? [])?.forEach(websocket => {
                                                        websocket.send(JSON.stringify({ event: 'editedActivity', id: activityId, newActivity: newActivity }));
                                                    });
                                                });
                                                JSON.parse(oldActivity.receiver).filter((oldReceiver: string) => !newActivity.receiver.includes(oldReceiver)).forEach((oldReceiver: string) => {
                                                    Array.from(websockets.get(res.locals.school)?.get(oldReceiver)?.values() ?? [])?.forEach(websocket => {
                                                        websocket.send(JSON.stringify({ event: 'deletedActivity', id: activityId }));
                                                    });
                                                });
                                                res.send({});
                                            } else {
                                                console.log(err);
                                                res.status(500).send({ error: "Server error." });
                                            }
                                        });
                                    } else {
                                        res.status(400).send({ error: "Missing required argument." });
                                    }
                                } else {
                                    res.status(403).send({ error: "Not authorized." });
                                }
                            } else {
                                res.status(403).send({ error: "Not authorized." });
                            }
                        } else {
                            res.status(403).send({ error: "Not authorized." });
                        }
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
    });

    app.delete('/activities/*', (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const activityId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        database.query(`SELECT * FROM activities`, async (err, dbRes) => {
            if (!err) {
                const activity = dbRes.rows.find(x => x.id === activityId);
                database.query(`SELECT * FROM users`, async (err, dbResu) => {
                    if (!err) {
                        if (activity.author === res.locals.user || JSON.parse(dbResu.rows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                            database.query(`DELETE FROM activities WHERE id = $1`, [activityId], async (err, dbRes) => {
                                if (!err) {
                                    dbResu.rows.filter(x => JSON.parse(x.administrator).includes(res.locals.school)).map(x => x.id).concat([...JSON.parse(activity.receiver), activity.author]).forEach((oldReceiver: string) => {
                                        Array.from(websockets.get(res.locals.school)?.get(oldReceiver)?.values() ?? [])?.forEach(websocket => {
                                            websocket.send(JSON.stringify({ event: 'deletedActivity', id: activityId }));
                                        });
                                    });
                                    res.send({});
                                } else {
                                    console.log(err);
                                    res.status(500).send({ error: "Server error." });
                                }
                            });
                        } else {
                            res.status(403).send({ error: "Not authorized." });
                        }
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
    });

    app.post('/activities/view/*', (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const activityId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                if (!JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).teacher)[res.locals.school] && !JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {

                    database.query(`SELECT * FROM activities`, async (err, dbRes) => {
                        if (!err) {
                            const activity = dbRes.rows.find(x => x.id === activityId);
                            if (activity) {
                                let viewed = JSON.parse(activity.viewed);
                                if (!viewed[res.locals.user]) {
                                    viewed[res.locals.user] = true;
                                    if (!err) {
                                        database.query(`UPDATE activities SET viewed = $1 WHERE id = $2`, [viewed, activityId], (err, dbRes) => {
                                            if (!err) {
                                                Array.from(websockets.get(res.locals.school)?.get(res.locals.user)?.values() ?? [])?.forEach(websocket => {
                                                    websocket.send(JSON.stringify({ event: 'viewedActivity', id: activityId }));
                                                });
                                                Array.from(websockets.get(res.locals.school)?.get(activity.author)?.values() ?? [])?.forEach(websocket => {
                                                    websocket.send(JSON.stringify({ event: 'viewedActivity', id: activityId, user: res.locals.user }));
                                                });
                                                res.send({});
                                            } else {
                                                console.log(err);
                                                res.status(500).send({ error: "Server error." });
                                            }
                                        });
                                    } else {
                                        console.log(err);
                                        res.status(500).send({ error: "Server error." });
                                    }
                                } else {
                                    res.status(400).send({ error: "Missing required argument." });
                                }
                            } else {
                                res.status(404).send({ error: "Not found." });
                            }
                        } else {
                            console.log(err);
                            res.status(500).send({ error: "Server error." });
                        }
                    });
                } else {
                    res.status(403).send({ error: "Not authorized." });
                }
            } else {
                console.log(err);
                res.status(500).send({ error: "Server error." });
            }
        });
    });

    app.post('/activities/deliver/*', (req: express.Request, res: express.Response) => {
        const files = fs.readdirSync(__dirname + '/../../files');

        const urlParamsValues: string[] = Object.values(req.params);
        const activityId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                const userRows = dbRes.rows;
                if (!JSON.parse(userRows.find(x => x.id === res.locals.user).teacher)[res.locals.school] && !JSON.parse(userRows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                    database.query(`SELECT * FROM activities`, async (err, dbRes) => {
                        if (!err) {
                            const activity = dbRes.rows.find(x => x.id === activityId);
                            if (activity) {
                                if (activity.expiration === 'false' || Date.now() <= Number(activity.expiration)) {
                                    if (!JSON.parse(activity.result)[res.locals.user]) {
                                        if (!(req.body.files ?? []).map((x: File) => !!files.find((y: string) => y.startsWith(x.id))).includes(false)) {
                                            let results = JSON.parse(activity.result);
                                            results[res.locals.user] = 'Unchecked';

                                            let delivered = JSON.parse(activity.delivered);
                                            delivered[res.locals.user] = {
                                                comments: req.body.comments ?? '',
                                                files: req.body.files ?? [],
                                                date: Date.now()
                                            };
                                            database.query(`UPDATE activities SET delivered = $1, result = $2 WHERE id = $3`, [JSON.stringify(delivered), JSON.stringify(results), activityId], (err, dbRes) => {
                                                if (!err) {

                                                    let websocketiedDelivery = delivered[res.locals.user];
                                                    websocketiedDelivery.name = userRows.find(y => y?.id === res.locals.user)?.name ?? "Deleted user";
                                                    Array.from(websockets.get(res.locals.school)?.get(activity.author)?.values() ?? [])?.forEach(websocket => {
                                                        websocket.send(JSON.stringify({ event: 'deliveredActivity', id: activityId, user: res.locals.user, delivery: websocketiedDelivery }));
                                                    });
                                                    res.send({});
                                                } else {
                                                    console.log(err);
                                                    res.status(500).send({ error: "Server error." });
                                                }
                                            });
                                        } else {
                                            res.status(400).send({})
                                        }
                                    } else {
                                        res.status(403).send({ error: "Not authorized." });
                                    }
                                } else {
                                    res.status(403).send({ error: "Not authorized." });
                                }
                            } else {
                                res.status(404).send({})
                            }
                        } else {
                            console.log(err);
                            res.status(500).send({ error: "Server error." });
                        }
                    });
                } else {
                    res.status(403).send({ error: "Not authorized." });
                }

            } else {
                console.log(err);
                res.status(500).send({ error: "Server error." });
            }
        });
    });

    app.post('/activities/result/*/*', (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const urlParams = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            });
        const activityId = urlParams[0];
        const userId = urlParams[1];

        database.query(`SELECT * FROM activities`, async (err, dbRes) => {
            if (!err) {
                const activity = dbRes.rows.find(x => x.id === activityId);
                if (activity?.author === res.locals.user && JSON.parse(activity.result)[userId] === 'Unchecked') {
                    if (req.body.result === 'Accepted' || req.body.result === 'Rejected') {
                        let results = JSON.parse(activity.result);
                        results[userId] = req.body.result;

                        database.query(`UPDATE activities SET result = $1 WHERE id = $2`, [JSON.stringify(results), activityId], (err, dbRes) => {
                            if (!err) {
                                Array.from(websockets.get(res.locals.school)?.get(res.locals.user)?.values() ?? [])?.forEach(websocket => {
                                    websocket.send(JSON.stringify({ event: 'resultActivity', id: activityId, user: userId, result: req.body.result }));
                                });
                                Array.from(websockets.get(res.locals.school)?.get(userId)?.values() ?? [])?.forEach(websocket => {
                                    websocket.send(JSON.stringify({ event: 'resultActivity', id: activityId, result: req.body.result }));
                                });
                                res.send({});
                            } else {
                                console.log(err);
                                res.status(500).send({ error: "Server error." });
                            }
                        });
                    } else {
                        res.status(400).send({ error: "Missing required argument." });
                    }
                } else {
                    res.status(403).send({ error: "Not authorized." });
                }
            } else {
                console.log(err);
                res.status(500).send({ error: "Server error." });
            }
        });
    });
};