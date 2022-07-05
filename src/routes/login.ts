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
                            database.query(`INSERT INTO users (token, id, email, number, tfa, name, grades, password, administrator, teacher, parent, schools, pending) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`, ['NEW', crypto.randomUUID(), req.body.email, '', '', req.body.name, '{}', await argon2.hash(req.body.password, { type: argon2.argon2id }), '[]', '{}', '{}', '[]', '[]'], (err, dbRes) => {
                                if (!err) {
                                    res.send({});
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

    app.post('/login', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                const user = dbRes.rows.find(x => x.email === req.body.email);
                if (user) {
                    try {
                        if (await argon2.verify(user.password, req.body.password, { type: argon2.argon2id })) {
                            database.query(`SELECT * FROM schools`, async (err, dbRes) => {
                                if (!err) {
                                    const verifiedUser = await checkLogin(true, user.token);
                                    if (verifiedUser.id !== "") {
                                        let realUser: User = user;
                                        delete realUser.password;
                                        delete realUser.administrator;
                                        delete realUser.teacher;
                                        delete realUser.parent;
                                        delete realUser.grades;
                                        realUser.schools = JSON.parse(user.schools).map((x: string) => dbRes.rows.find(y => y.id === x));
                                        res.send(realUser);
                                    } else {
                                        const userPrivateless: User = user;
                                        delete userPrivateless.token;
                                        delete userPrivateless.password;
                                        delete userPrivateless.administrator;
                                        delete userPrivateless.teacher;
                                        delete userPrivateless.parent;
                                        delete userPrivateless.grades;
                                        userPrivateless.schools = JSON.parse(user.schools).map((x: string) => dbRes.rows.find(y => y.id === x));
                                        const newToken = "Bearer " + (await generateToken(userPrivateless));
                                        database.query(`UPDATE users SET token = '${newToken}' WHERE id = '${user.id}'`, err => {
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
                    res.status(401).send({ error: "Invalid credentials." });
                }
            } else {
                console.log(err);
                res.status(500).send({ error: "Server error." });
            }
        });
    });

    app.post('/loginByToken', async (req: express.Request, res: express.Response) => {
        const user = await checkLogin(true, req.body.token);
        const valid = user.token === req.body.token;
        if (valid) {
            database.query(`SELECT * FROM schools`, (err, dbRes) => {
                if (!err) {
                    let realUser: User = user;
                    delete realUser.password;
                    delete realUser.administrator;
                    delete realUser.teacher;
                    delete realUser.parent;
                    delete realUser.grades;
                    realUser.schools = JSON.parse(user.schools).map((x: string) => dbRes.rows.find(y => y.id === x));
                    realUser.pending = JSON.parse(user.pending).map((x: string) => dbRes.rows.find(y => y.id === x));
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
            .setIssuer('school')
            .setAudience('school')
            .setExpirationTime('7d')
            .sign(privateKey);
    }

};