import { File, Receiver } from '../interfaces';

import express from "express";
import { Client } from 'pg';
import crypto from 'crypto';
import fs from 'fs';

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, Map<string, WebSocket>>>) => {
    app.get('/messages', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM messages`, async (err, dbRes) => {
            if (!err) {
                let messages = dbRes.rows.filter(x => x.school === res.locals.school);
                database.query(`SELECT * FROM users`, async (err, dbRes) => {
                    if (!err) {
                        if (!JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                            messages = messages.filter(x => x.receiver.includes(res.locals.user) || x.author === res.locals.user);
                        }
                        res.send(messages.map(x => {
                            try {
                                const pdf = JSON.parse(x.content);
                                delete x.content;
                                x.pdf = pdf.pdf;
                            } catch { }
                            x.files = JSON.parse(x.files);
                            x.receiver = JSON.parse(x.receiver).map((x: string) => { return { id: x, name: dbRes.rows.find(y => y?.id === x)?.name ?? "Deleted user" } });
                            x.author = { id: x.author, name: dbRes.rows.find(y => y?.id === x.author)?.name ?? "Deleted user" };
                            x.date = Number(x.date);
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

    app.post('/messages', (req: express.Request, res: express.Response) => {
        const files = fs.readdirSync(__dirname + '/../../files');
        const message = {
            id: crypto.randomUUID(),
            title: req.body.title,
            content: req.body.content,
            files: req.body.files ?? [],
            author: res.locals.user,
            date: Date.now(),
            receiver: req.body.receiver ?? [],
            school: res.locals.school
        };
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                let avaliableUsers = dbRes.rows;
                const user = dbRes.rows.find(x => x.id === res.locals.user);
                if (!JSON.parse(user.teacher)[res.locals.school] && !JSON.parse(user.administrator)[res.locals.school]) {
                    avaliableUsers = avaliableUsers.filter(x => JSON.parse(x.teacher)[res.locals.school] || JSON.parse(x.administrator).includes(res.locals.school));
                }
                avaliableUsers = avaliableUsers.map(x => x.id);
                if (!message.receiver.some((x: string) => !avaliableUsers.includes(x))) {
                    if (message.title && message.content && message.receiver.length > 0 && !message.files.map((x: File) => !!files.find((y: string) => y.startsWith(x.id))).includes(false)) {
                        const users = dbRes.rows;
                        database.query(`INSERT INTO messages (id, title, content, files, author, date, receiver, school) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`, [message.id, message.title, !message.content?.pdf ? message.content : JSON.stringify(message.content), JSON.stringify(message.files), message.author, message.date.toString(), JSON.stringify(message.receiver), message.school], (err, dbRes) => {
                            if (!err) {
                                let websocketiedMessage = { ...message, pdf: message.content?.pdf };
                                websocketiedMessage.receiver = websocketiedMessage.receiver.map((x: string) => { return { id: x, name: users.find(y => y?.id === x)?.name ?? "Deleted user" } });
                                websocketiedMessage.author = { id: websocketiedMessage.author, name: users.find(y => y?.id === websocketiedMessage.author)?.name ?? "Deleted user" };
                                message.receiver.forEach((receiver: string) => {
                                    Array.from(websockets.get(res.locals.school)?.get(receiver)?.values() ?? [])?.forEach(websocket => {
                                        websocket.send(JSON.stringify({ event: 'newMessage', ...websocketiedMessage }));
                                    });
                                });
                                Array.from(websockets.get(res.locals.school)?.get(message.author)?.values() ?? [])?.forEach(websocket => {
                                    websocket.send(JSON.stringify({ event: 'newMessage', ...websocketiedMessage }));
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
                console.log(err);
                res.status(500).send({ error: "Server error." });
            }
        });
    });

    app.patch('/messages/*', (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const messageId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        const files = fs.readdirSync(__dirname + '/../../files');
        database.query(`SELECT * FROM messages`, async (err, dbRes) => {
            if (!err) {
                const oldMessage = dbRes.rows.find(x => x.id === messageId);
                let pdf = false;
                try {
                    const pdfGet = JSON.parse(oldMessage.content);
                    if (pdfGet.pdf) {
                        pdf = true;
                    }
                } catch { }
                const newMessage = {
                    title: req.body.title,
                    content: pdf ? oldMessage.content : req.body.content,
                    files: req.body.files ?? [],
                    receiver: req.body.receiver ?? []
                };
                database.query(`SELECT * FROM users`, async (err, dbResu) => {
                    if (!err) {
                        if (oldMessage?.author === res.locals.user || JSON.parse(dbResu.rows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                            let avaliableUsers = dbRes.rows;
                            const user = dbRes.rows.find(x => x.id === res.locals.user);
                            if (!JSON.parse(user.teacher)[res.locals.school] && !JSON.parse(user.administrator)[res.locals.school]) {
                                avaliableUsers = avaliableUsers.filter(x => JSON.parse(x.teacher)[res.locals.school] || JSON.parse(x.administrator).includes(res.locals.school));
                }
                avaliableUsers = avaliableUsers.map(x => x.id);
                            if (!newMessage.receiver.some((x: string) => !avaliableUsers.includes(x))) {
                                if (newMessage.title && (oldMessage.pdf || newMessage.content) && newMessage.receiver.length > 0 && !newMessage.files.map((x: File) => !!files.find((y: string) => y.startsWith(x.id))).includes(false)) {
                                    const users = dbRes.rows;
                                    database.query(`UPDATE messages SET title = $1, content = $2, files = $3, receiver = $4 WHERE id = $5`, [newMessage.title, newMessage.content, JSON.stringify(newMessage.files), JSON.stringify(newMessage.receiver), messageId], (err, dbRes) => {
                                        if (!err) {
                                            let websocketiedMessage = { ...newMessage };
                                            websocketiedMessage.receiver = websocketiedMessage.receiver.map((x: string) => { return { id: x, name: users.find(y => y?.id === x)?.name ?? "Deleted user" } });
                                            newMessage.receiver.forEach((receiver: string) => {
                                                Array.from(websockets.get(res.locals.school)?.get(receiver)?.values() ?? [])?.forEach(websocket => {
                                                    websocket.send(JSON.stringify({ event: 'editedMessage', id: messageId, newMessage: websocketiedMessage }));
                                                });
                                            });
                                            JSON.parse(oldMessage.receiver).filter((oldReceiver: string) => !newMessage.receiver.includes(oldReceiver)).forEach((oldReceiver: string) => {
                                                Array.from(websockets.get(res.locals.school)?.get(oldReceiver)?.values() ?? [])?.forEach(websocket => {
                                                    websocket.send(JSON.stringify({ event: 'deletedMessage', id: messageId }));
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

    app.delete('/messages/*', (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const messageId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        database.query(`SELECT * FROM messages`, async (err, dbRes) => {
            if (!err) {
                const message = dbRes.rows.find(x => x.id === messageId);
                database.query(`SELECT * FROM users`, async (err, dbResu) => {
                    if (!err) {
                        if (message?.author === res.locals.user || JSON.parse(dbResu.rows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                            database.query(`DELETE FROM messages WHERE id = $1`, [messageId], async (err, dbRes) => {
                                if (!err) {
                                    JSON.parse(message.receiver).forEach((oldReceiver: string) => {
                                        Array.from(websockets.get(res.locals.school)?.get(oldReceiver)?.values() ?? [])?.forEach(websocket => {
                                            websocket.send(JSON.stringify({ event: 'deletedMessage', id: messageId }));
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
};