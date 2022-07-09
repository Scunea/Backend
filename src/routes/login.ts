import { User } from '../interfaces';

import express from "express";
import argon2 from 'argon2';
import * as jose from 'jose'
import { Client } from 'pg';
import fs from 'fs';
import crypto from 'crypto';

export default (app: express.Application, database: Client, checkLogin: any) => {
    app.post('/signup', async (req: express.Request, res: express.Response) => {
        if (req.body.email && req.body.name && req.body.password) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    if (!dbRes.rows.find(x => x.email === req.body.email)) {
                        if (/^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/.test(req.body.email)) {
                            database.query(`INSERT INTO users (token, id, email, verified, verificator, tfa, name, grades, password, administrator, teacher, parents, pendingparents, schools, pendingschools) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`, ['NEW', crypto.randomUUID(), req.body.email, false, crypto.randomUUID(), '', req.body.name, '{}', await argon2.hash(req.body.password, { type: argon2.argon2id }), '[]', '{}', '[]', '[]', '[]', '[]'], (err, dbRes) => {
                                if (!err) {
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
        } else {
            res.status(400).send({ error: "Missing required argument." });
        }
    });

    app.post('/verify/*', (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const otp = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];


        if (req.body.email && req.body.password && otp) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    const user = dbRes.rows.find(x => x.email === req.body.email);
                    if (user) {
                        if (!user.verifieed && user.verificator === otp) {
                            try {
                                if (await argon2.verify(user.password, req.body.password, { type: argon2.argon2id })) {
                                    database.query(`SELECT * FROM schools`, async (err, dbRes) => {
                                        if (!err) {
                                            const verifiedUser = await checkLogin(true, user.token);
                                            if (verifiedUser.id !== "") {
                                                let realUser: User = user;
                                                delete realUser.password;
                                                delete realUser.verified;
                                                delete realUser.verificator;
                                                realUser.tfa = Boolean(realUser.tfa);
                                                delete realUser.administrator;
                                                delete realUser.teacher;
                                                delete realUser.parents;
                                                delete realUser.pendingparents;
                                                delete realUser.grades;
                                                realUser.schools = JSON.parse(user.schools).map((x: string) => dbRes.rows.find(y => y.id === x));
                                                realUser.pendingschools = JSON.parse(user.pendingschools).map((x: string) => dbRes.rows.find(y => y.id === x));
                                                res.send(realUser);
                                            } else {
                                                const userPrivateless: User = user;
                                                delete userPrivateless.token;
                                                delete userPrivateless.password;
                                                delete userPrivateless.verified;
                                                delete userPrivateless.verificator;
                                                userPrivateless.tfa = Boolean(userPrivateless.tfa);
                                                delete userPrivateless.administrator;
                                                delete userPrivateless.teacher;
                                                delete userPrivateless.parents;
                                                delete userPrivateless.pendingparents;
                                                delete userPrivateless.grades;
                                                userPrivateless.schools = JSON.parse(user.schools).map((x: string) => dbRes.rows.find(y => y.id === x));
                                                userPrivateless.pendingschools = JSON.parse(user.pendingschools).map((x: string) => dbRes.rows.find(y => y.id === x));
                                                const newToken = "Bearer " + (await generateToken(userPrivateless));
                                                database.query('UPDATE users SET token = $1, verified = $2, verificator = $3 WHERE id = $3', [newToken, true, '', user.id], err => {
                                                    if (!err) {
                                                        let newUser = user;
                                                        user.token = newToken;
                                                        res.send(newUser);
                                                    } else {
                                                        console.log(err);
                                                        res.status(500).send({ error: "Server error." });
                                                    }
                                                });
                                            }
                                        } else {
                                            console.log(err);
                                            res.status(500).send({ error: "Server error." });
                                        }
                                    });
                                } else {
                                    res.status(401).send({ error: "Invalid credentials." });
                                }
                            } catch (e) {
                                console.log(err);
                                res.status(500).send({ error: "Server error." });
                            }
                        } else {
                            res.status(403).send({ error: "Not authorized." });
                        }
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

    app.post('/login', (req: express.Request, res: express.Response) => {
        if (req.body.email && req.body.password) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    const user = dbRes.rows.find(x => x.email === req.body.email);
                    if (user) {
                        if (user.verified) {
                            try {
                                if (await argon2.verify(user.password, req.body.password, { type: argon2.argon2id })) {
                                    if (!user.tfa || (user.tfa && req.body.otp)) {
                                        database.query(`SELECT * FROM schools`, async (err, dbRes) => {
                                            if (!err) {
                                                const verifiedUser = await checkLogin(true, user.token);
                                                if (verifiedUser.id !== "") {
                                                    let realUser: User = user;
                                                    delete realUser.password;
                                                    delete realUser.verified;
                                                    delete realUser.verificator;
                                                    realUser.tfa = Boolean(realUser.tfa);
                                                    delete realUser.administrator;
                                                    delete realUser.teacher;
                                                    delete realUser.parents;
                                                    delete realUser.pendingparents;
                                                    delete realUser.grades;
                                                    realUser.schools = JSON.parse(user.schools).map((x: string) => dbRes.rows.find(y => y.id === x));
                                                    realUser.pendingschools = JSON.parse(user.pendingschools).map((x: string) => dbRes.rows.find(y => y.id === x));
                                                    res.send(realUser);
                                                } else {
                                                    const userPrivateless: User = user;
                                                    delete userPrivateless.token;
                                                    delete userPrivateless.password;
                                                    delete userPrivateless.verified;
                                                    delete userPrivateless.verificator;
                                                    userPrivateless.tfa = Boolean(userPrivateless.tfa);
                                                    delete userPrivateless.administrator;
                                                    delete userPrivateless.teacher;
                                                    delete userPrivateless.parents;
                                                    delete userPrivateless.pendingparents;
                                                    delete userPrivateless.grades;
                                                    userPrivateless.schools = JSON.parse(user.schools).map((x: string) => dbRes.rows.find(y => y.id === x));
                                                    userPrivateless.pendingschools = JSON.parse(user.pendingschools).map((x: string) => dbRes.rows.find(y => y.id === x));
                                                    const newToken = "Bearer " + (await generateToken(userPrivateless));
                                                    database.query('UPDATE users SET token = $1 WHERE id = $2', [newToken, user.id], err => {
                                                        if (!err) {
                                                            let newUser = user;
                                                            user.token = newToken;
                                                            res.send(newUser);
                                                        } else {
                                                            console.log(err);
                                                            res.status(500).send({ error: "Server error." });
                                                        }
                                                    });
                                                }
                                            } else {
                                                console.log(err);
                                                res.status(500).send({ error: "Server error." });
                                            }
                                        });
                                    } else {
                                        res.status(400).send({ missingOtp: true });
                                    }
                                } else {
                                    res.status(401).send({ error: "Invalid credentials." });
                                }
                            } catch (e) {
                                console.log(err);
                                res.status(500).send({ error: "Server error." });
                            }
                        } else {
                            res.status(403).send({ error: "Not authorized." });
                        }
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

    app.post('/loginByToken', async (req: express.Request, res: express.Response) => {
        const user = await checkLogin(true, req.body.token);
        const valid = user.token === req.body.token;
        if (valid) {
            database.query(`SELECT * FROM schools`, (err, dbRes) => {
                if (!err) {
                    let realUser: User = user;
                    delete realUser.password;
                    delete realUser.verified;
                    delete realUser.verificator;
                    realUser.tfa = Boolean(realUser.tfa);
                    delete realUser.administrator;
                    delete realUser.teacher;
                    delete realUser.parents;
                    delete realUser.pendingparents;
                    delete realUser.grades;
                    realUser.schools = JSON.parse(user.schools).map((x: string) => dbRes.rows.find(y => y.id === x));
                    realUser.pendingschools = JSON.parse(user.pendingschools).map((x: string) => dbRes.rows.find(y => y.id === x));
                    res.send(realUser);
                }
            });
        } else {
            res.status(401).send({ error: "Invalid credentials." });
        }
    });

    async function generateToken(info: User) {
        const privateKey = await jose.importPKCS8(fs.readFileSync(__dirname + '/../../private.key').toString(), 'ES256');
        return await new jose.SignJWT({ info })
            .setProtectedHeader({ alg: 'ES256' })
            .setIssuedAt()
            .setIssuer('scunea')
            .setAudience('scunea')
            .setExpirationTime('7d')
            .sign(privateKey);
    }

};