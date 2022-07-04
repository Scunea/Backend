import express from "express";
import { Grade, GradeParsed } from "interfaces";
import { Client } from 'pg';

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, WebSocket[]>>) => {
    app.get('/grades', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                const user = dbRes.rows.find(x => x.id === res.locals.user);
                if (JSON.parse(user.teacher)[res.locals.school]) {
                    const students = dbRes.rows.filter(x => JSON.parse(x.schools).includes(res.locals.school) && !JSON.parse(x.administrator).includes(res.locals.school) && !JSON.parse(x.teacher)[res.locals.school] && !dbRes.rows.find(y => JSON.parse(y.parent)[res.locals.school]?.includes(x.id)));
                    res.send(students.map(x => {
                        const xUser = (JSON.parse(x.grades)[res.locals.school] ?? {})[user.id];
                        return {
                            id: x.id,
                            fullName: x.name,
                            subject: JSON.parse(user.teacher)[res.locals.school],
                            deliberation: xUser?.deliberation,
                            conceptual: xUser?.conceptual,
                            averageFirstFour: xUser?.averageFirstFour,
                            averageSecondFour: xUser?.averageSecondFour,
                            final: xUser?.final
                        }
                    }));
                } else {
                    res.status(403).send({ error: "Not authorized." });
                }
            }
        });
    });

    app.get('/grades/*', (req: express.Request, res: express.Response) => {
        const urlParamsValues: string[] = Object.values(req.params);
        const userId = urlParamsValues
            .map((x) => x.replace(/\//g, ''))
            .filter((x) => {
                return x !== '';
            })[0];

        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                const user = dbRes.rows.find(x => x.id === res.locals.user);
                if (JSON.parse(user.administrator).includes(res.locals.school)) {
                    const grades = JSON.parse(dbRes.rows.find(x => x.id === userId).grades)[res.locals.school] ?? {};
                    res.send(Object.keys(grades).map(x => {
                        return {
                            subject: JSON.parse(dbRes.rows.find(y => y.id === x).teacher)[res.locals.school],
                            deliberation: grades[x].deliberation,
                            conceptual: grades[x].conceptual,
                            averageFirstFour: grades[x].averageFirstFour,
                            averageSecondFour: grades[x].averageSecondFour,
                            final: grades[x].final
                        }
                    }));
                } else {
                    res.status(403).send({ error: "Not authorized." });
                }
            }
        });
    });

    app.post('/grades', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                const user = dbRes.rows.find(x => x.id === res.locals.user);
                if ((JSON.parse(user.teacher)[res.locals.school] && !req.body.find((x: Grade) => x.subject !== JSON.parse(user.teacher)[res.locals.school])) || JSON.parse(user.administrator).includes(res.locals.school)) {
                    const students = dbRes.rows.filter(x => JSON.parse(x.schools).includes(res.locals.school) && !JSON.parse(x.administrator).includes(res.locals.school) && !JSON.parse(x.teacher)[res.locals.school] && !dbRes.rows.find(y => JSON.parse(y.parent)[res.locals.school]?.includes(x.id)));
                    if (!students.some(x => !req.body.map((x: Grade) => x.id).includes(x.id)) || !req.body.map((x: Grade) => x.id).some((x: string) => !students.map(x => x.id).includes(x))) {
                        req.body.forEach((x: Grade) => {
                            const parsed: GradeParsed = {
                                deliberation: x.deliberation,
                                conceptual: x.conceptual,
                                averageFirstFour: x.averageFirstFour,
                                averageSecondFour: x.averageSecondFour,
                                final: x.final
                            };
                            const extraParse = JSON.parse(dbRes.rows.find(y => y.id === x.id).grades);
                            if (!extraParse[res.locals.school]) {
                                extraParse[res.locals.school] = {};
                            }
                            extraParse[res.locals.school][user.id] = parsed;
                            database.query(`UPDATE users SET grades = $1 WHERE id = $2`, [extraParse, x.id], (err, dbReso) => {
                                if (!err) {
                                    students.forEach(student => {
                                        websockets.get(res.locals.school)?.get(student.id)?.forEach(websocket => {
                                            const grades = JSON.parse(student.grades)[res.locals.school] ?? {};
                                            websocket.send(JSON.stringify({
                                                event: 'newGrades', grades: Object.keys(grades).map(x => {
                                                    return {
                                                        subject: JSON.parse(user.teacher)[res.locals.school],
                                                        deliberation: grades[x].deliberation,
                                                        conceptual: grades[x].conceptual,
                                                        averageFirstFour: grades[x].averageFirstFour,
                                                        averageSecondFour: grades[x].averageSecondFour,
                                                        final: grades[x].final
                                                    }
                                                })
                                            }));
                                        });
                                    });
                                    res.status(200).send();
                                } else {
                                    console.log(err);
                                    res.status(500).send({ error: "Server error." });
                                }
                            });
                        });
                    } else {
                        res.status(400).send({ error: "Missing required argument." });
                    }
                } else {
                    res.status(403).send({ error: "Not authorized." });
                }
            }
        });
    });
};