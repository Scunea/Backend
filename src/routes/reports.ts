import express from "express";
import { Client } from 'pg';
import crypto from 'crypto';
import fs from 'fs';

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, WebSocket[]>>) => {
    app.get('/reports', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM reports`, async (err, dbRes) => {
            if (!err) {
                const reports = dbRes.rows.filter(x => x.school === res.locals.school);
                database.query(`SELECT * FROM users`, async (err, dbRes) => {
                    if (!err) {
                        res.send(reports.map(x => {
                            x.file = JSON.parse(x.file);
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

    app.post('/reports', (req: express.Request, res: express.Response) => {
        const files = fs.readdirSync(__dirname + '/../../files');
        const report = {
            id: crypto.randomUUID(),
            title: req.body.title,
            file: req.body.file ?? [],
            author: res.locals.user,
            date: Date.now(),
            school: res.locals.school
        };
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                if (JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                    if (report.title && report.file && files.includes(report.file.id)) {
                        const users = dbRes.rows;
                        database.query(`INSERT INTO reports (id, title, file, author, date, school) VALUES($1, $2, $3, $4, $5, $6)`, [report.id, report.title, JSON.stringify(report.file), report.author, report.date.toString(), report.school], (err, dbRes) => {
                            if (!err) {
                                let websocketiedReport = { ...report };
                                websocketiedReport.author = { id: websocketiedReport.author, name: users.find(y => y?.id === websocketiedReport.author)?.name ?? "Deleted user" };
                                users.map(x => x.id).forEach(receiver => {
                                    websockets.get(res.locals.school)?.get(receiver)?.forEach(websocket => {
                                        websocket.send(JSON.stringify({ event: 'newReport', ...websocketiedReport }));
                                    });
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

    app.patch('/reports/*', (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const reportId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        database.query(`SELECT * FROM reports`, async (err, dbRes) => {
            if (!err) {
                const oldReport = dbRes.rows.find(x => x.id === reportId);
                if (oldReport) {
                    database.query(`SELECT * FROM users`, async (err, dbRes) => {
                        if (!err) {
                            if (JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                                if (req.body.title) {
                                    const users = dbRes.rows;
                                    database.query(`UPDATE reports SET title = $1 WHERE id = $2`, [req.body.title, reportId], (err, dbRes) => {
                                        if (!err) {
                                            users.map(x => x.id).forEach(receiver => {
                                                websockets.get(res.locals.school)?.get(receiver)?.forEach(websocket => {
                                                    websocket.send(JSON.stringify({ event: 'editedReport', id: reportId, newTitle: req.body.title }));
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
    });

    app.delete('/reports/*', (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const reportId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                if (JSON.parse(dbRes.rows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                    const users = dbRes.rows;
                    database.query(`DELETE FROM reports WHERE id = $1`, [reportId], async (err, dbRes) => {
                        if (!err) {
                            users.map(x => x.id).forEach(receiver => {
                                websockets.get(res.locals.school)?.get(receiver)?.forEach(websocket => {
                                    websocket.send(JSON.stringify({ event: 'deletedReport', id: reportId }));
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
    });
};