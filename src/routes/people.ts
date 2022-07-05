import express from "express";
import { Client } from 'pg';
import argon2 from 'argon2';
import { User } from "interfaces";

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, WebSocket[]>>) => {
    app.get('/people', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                if (dbRes.rows.find(x => x.id === res.locals.user).administrator.includes(res.locals.school)) {
                    const schoolUsers = dbRes.rows.filter(user => JSON.parse(user.schools).includes(res.locals.school));
                    res.send(schoolUsers.map(user => {
                        const child = schoolUsers.find(x => JSON.parse(x.parent)[res.locals.school]?.includes(user.id));
                        let fixedChild;
                        if (child) {
                            fixedChild = {
                                id: child.id,
                                name: child.name
                            }
                        }
                        return {
                            id: user.id,
                            name: user.name,
                            type: JSON.parse(user.administrator).includes(res.locals.school) ?
                                'Administrator' : JSON.parse(user.teacher)[res.locals.school] ?
                                    'Teacher' : schoolUsers.find(x => JSON.parse(x.parent)[res.locals.school]?.includes(user.id)) ?
                                        'Parent' :
                                        'Student',
                            subject: JSON.parse(user.teacher)[res.locals.school],
                            child: fixedChild
                        };
                    }));
                } else {
                    res.status(403).send({ error: "Not authorized." });
                }
            } else {
                console.log(err);
                res.status(500).send({ error: "Server error." });
            }
        });
    });

    app.post('/people', async (req: express.Request, res: express.Response) => {
        if (req.body.id && req.body.name && req.body.password && req.body.type && (req.body.type === 'administrator' || (req.body.type === 'teacher' && req.body.subject) || req.body.type === 'student') || req.body.type === 'parent') {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (err) {
                    if (dbRes.rows.find(x => x.id === res.locals.user).administrator.includes(res.locals.school)) {
                        if (req.body.type !== 'parent' || (req.body.type === 'parent' && dbRes.rows.find(x => x.id === req.body.child))) {
                            const user = {
                                id: req.body.id,
                                name: req.body.name,
                                password: await argon2.hash(req.body.password, {
                                    type: argon2.argon2id
                                }),
                                subject: req.body.subject,
                                child: req.body.child,
                                type: req.body.type
                            }
                            database.query(`INSERT INTO users (token, id, name, grades, password, administrator, teacher, parent, schools) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`, ['NEW', user.id, user.name, '{}', user.password, (() => {
                                if (user.type === 'administrator') {
                                    return JSON.stringify([res.locals.school]);
                                } else {
                                    return '[]';
                                }
                            })(), (() => {
                                if (user.type === 'teacher') {
                                    return JSON.stringify({ [res.locals.school]: user.subject });
                                } else {
                                    return '[]';
                                }
                            })(), '[]', JSON.stringify([res.locals.school])], (err, dbResp) => {
                                if (!err) {

                                    if (req.body.type === 'parent') {
                                        let parents = JSON.parse(dbRes.rows.find(x => x.id === user.child).parent);
                                        if (Array.isArray(parents[res.locals.school])) {
                                            parents[res.locals.school].push(user.id);
                                        } else {
                                            parents[res.locals.school] = [user.id];
                                        }

                                        database.query(`UPDATE users SET parent = $1 WHERE id = $2`, [parents, user.child], (err, dbRes) => {
                                            if (!err) {
                                                dbRes.rows.forEach((userToSend: any) => {
                                                    if (user.type === 'teacher' || JSON.parse(userToSend.administrator).includes(res.locals.school)) {
                                                        websockets.get(res.locals.school)?.get(userToSend.id)?.forEach(websocket => {
                                                            websocket.send(JSON.stringify({ event: 'newUser', user: user }));
                                                        });
                                                    }
                                                });
                                                res.status(201).send({
                                                    id: user.id,
                                                    name: user.name,
                                                    subject: user.subject,
                                                    child: '',
                                                    type: user.type
                                                });
                                            } else {
                                                console.log(err);
                                                res.status(500).send({ error: "Server error." });
                                            }
                                        });
                                    } else {
                                        dbRes.rows.forEach((userToSend: any) => {
                                            if (user.type === 'teacher' || JSON.parse(userToSend.administrator).includes(res.locals.school)) {
                                                websockets.get(res.locals.school)?.get(userToSend.id)?.forEach(websocket => {
                                                    let privaterisedUser: any = user;
                                                    delete privaterisedUser.password;
                                                    websocket.send(JSON.stringify({ event: 'newUser', user: privaterisedUser }));
                                                });
                                            }
                                        });
                                        res.status(201).send({
                                            id: user.id,
                                            name: user.name,
                                            subject: user.subject,
                                            child: '',
                                            type: user.type
                                        });
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
                        res.status(403).send({ error: "Not authorized." });
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

    app.patch('/people', async (req: express.Request, res: express.Response) => {
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
                                        websockets.get(res.locals.school)?.get(userToSend.id)?.forEach(websocket => {
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

    app.patch('/people/*', async (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const userId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        if ((req.body.name || req.body.password || req.body.subject)) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    if (dbRes.rows.find(x => x.id === res.locals.user).administrator.includes(res.locals.school)) {

                        const user = dbRes.rows.find(x => x.id === userId);
                        if (user) {
                            let teacher = JSON.parse(user.teacher);
                            if (teacher[res.locals.school] && req.body.subject) {
                                teacher[res.locals.school] = req.body.subject;
                            }
                            user.name = req.body.name;
                            user.teacher = JSON.stringify(teacher);
                            if (req.body.password) {
                                user.password = await argon2.hash(req.body.password, { type: argon2.argon2id });
                            }
                            database.query(`UPDATE users SET name = $1, teacher = $2, password = $3 WHERE id = $4`, [user.name, user.teacher, user.password, user.id], (err, dbResp) => {
                                if (!err) {
                                    dbRes.rows.forEach((userToSend: any) => {
                                        if (JSON.parse(user.teacher)[res.locals.school] || JSON.parse(userToSend.administrator).includes(res.locals.school)) {
                                            const child = dbRes.rows.find(x => JSON.parse(x.parent)[res.locals.school]?.includes(user.id));
                                            let fixedChild: { id: string; name: string; };
                                            if (child) {
                                                fixedChild = {
                                                    id: child.id,
                                                    name: child.name
                                                }
                                            }
                                            websockets.get(res.locals.school)?.get(userToSend.id)?.forEach(websocket => {
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
                } else {
                    res.status(403).send({ error: "Not authorized." });
                }
            });
        } else {
            res.status(400).send({ error: "Missing required argument." });
        }
    });

    app.delete('/people', async (req: express.Request, res: express.Response) => {
        if (req.body.tos) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    if (dbRes.rows.find(x => x.id === res.locals.user).administrator.includes(res.locals.school)) {
                        const users = dbRes.rows.filter(x => req.body.tos.includes(x.id));
                        database.query(`DELETE FROM users WHERE id = ANY($1)`, [users.map(x => x.id)], (err, dbResp) => {
                            if (!err) {
                                dbRes.rows.forEach((userToSend: any) => {
                                    users.forEach(user => {
                                        if (JSON.parse(user.teacher)[res.locals.school] || JSON.parse(userToSend.administrator).includes(res.locals.school)) {
                                            websockets.get(res.locals.school)?.get(userToSend.id)?.forEach(websocket => {
                                                websocket.send(JSON.stringify({ event: 'deletedUser', userId: user.id }));
                                            });
                                        }
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
            res.status(400).send({ error: "Missing required argument." });
        }
    });
};