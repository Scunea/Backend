import express from "express";
import { Client } from 'pg';
import argon2 from 'argon2';
import * as twofactor from 'node-2fa';
import { User } from "interfaces";

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, Map<string, WebSocket>>>) => {
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
                                        const children = dbRes.rows.filter(x => JSON.parse(x.parents)?.includes(user.id));
                                        const fixedChildren = children.map((child: User) => {
                                            return {
                                                id: child.id,
                                                name: child.name
                                            }
                                        });
                                        websockets?.forEach(websocket => {
                                            if (websocket.has(userToSend.id)) {
                                                websocket.get(userToSend.id)?.forEach(websocket => {
                                                    websocket.send(JSON.stringify({
                                                        event: 'editedUser', user: {
                                                            id: user.id,
                                                            name: user.name,
                                                            subject: JSON.parse(user.teacher)[res.locals.school],
                                                            children: fixedChildren,
                                                            type: JSON.parse(user.administrator).includes(res.locals.school) ?
                                                                'Administrator' : JSON.parse(user.teacher)[res.locals.school] ?
                                                                    'Teacher' : dbRes.rows.find(x => JSON.parse(x.schools).includes(res.locals.school) && JSON.parse(x.parents)?.includes(user.id)) ?
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

    app.delete('/account', async (req: express.Request, res: express.Response) => {
        if (req.body.password) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                const user = dbRes.rows.find(x => x.id === res.locals.user);
                        if (!user.tfa || (user.tfa && twofactor.verifyToken(user.tfa, req.body.otp ?? ''))) {
                            if (await argon2.verify(user.password, req.body.password, { type: argon2.argon2id })) {
                                database.query('DELETE FROM users WHERE id = $1', [user.id], async (err, dbRes) => {
                                    if(!err) {    
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
                            res.status(401).send({ error: "Invalid credentials." });
                        }
                
            });
        } else {
            res.status(400).send({ error: "Missing required argument." });
        }
    });

    app.post('/otp', async (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                const user = dbRes.rows.find(x => x.id === res.locals.user);
                    if (!user.tfa) {
                        const secret = twofactor.generateSecret({ name: 'Scunea', account: user.email });;
                        res.send(secret);
                    } else {
                        res.status(403).send({ error: "Not authorized." });;
                    }
                } else {
                    console.log(err);
                    res.status(500).send({ error: "Server error." });
                }
        });
    });

    app.post('/otp/*', async (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const otpCode = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        if (req.body.password && otpCode) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
            
                const user = dbRes.rows.find(x => x.id === res.locals.user);
                        if (!user.tfa) {
                            if (await argon2.verify(user.password, req.body.password, { type: argon2.argon2id }) && twofactor.verifyToken(otpCode, req.body.otp)) {
                                database.query('UPDATE users SET tfa = $1 WHERE id = $2', [otpCode, user.id], async (err, dbRes) => {
                                    if(!err) {    
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
                            res.status(403).send({ error: "Not authorized." });;
                        }
                
            });
        } else {
            res.status(400).send({ error: "Missing required argument." });
        }
    });

    app.delete('/otp', async (req: express.Request, res: express.Response) => {
        if (req.body.password && req.body.otp) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
            
                const user = dbRes.rows.find(x => x.id === res.locals.user);
                        if (user.tfa) {
                            if (await argon2.verify(user.password, req.body.password, { type: argon2.argon2id }) && twofactor.verifyToken(user.tfa, req.body.otp)) {
                                database.query('UPDATE users SET tfa = $1 WHERE id = $2', ['', user.id], async (err, dbRes) => {
                                    if(!err) {    
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
                            res.status(403).send({ error: "Not authorized." });;
                        }
                
            });
        } else {
            res.status(400).send({ error: "Missing required argument." });
        }
    });
};