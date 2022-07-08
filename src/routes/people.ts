import express from "express";
import { User } from "interfaces";
import { Client } from 'pg';

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, Map<string, WebSocket>>>) => {
    app.get('/people', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                if (dbRes.rows.find(x => x.id === res.locals.user).administrator.includes(res.locals.school)) {
                    const schoolUsers = dbRes.rows.filter(user => JSON.parse(user.schools).includes(res.locals.school) || JSON.parse(user.pendingschools).includes(res.locals.school));
                    res.send(schoolUsers.map(user => {
                        const children = dbRes.rows.filter(x => JSON.parse(x.parents)?.includes(user.id));
                                        const fixedChildren = children.map((child: User) => {
                                            return {
                                                id: child.id,
                                                name: child.name
                                            }
                                        });
                                        
                        return {
                            id: user.id,
                            name: user.name,
                            email: user.email,
                            type: !JSON.parse(user.pendingschools).includes(res.locals.school) ? JSON.parse(user.administrator).includes(res.locals.school) ?
                                'Administrator' : JSON.parse(user.teacher)[res.locals.school] ?
                                    'Teacher' : schoolUsers.find(x => JSON.parse(x.schools).includes(res.locals.school) && JSON.parse(x.parents)?.includes(user.id)) ?
                                        'Parent' :
                                        'Student' : 'Pending',
                            subject: !JSON.parse(user.pendingschools).includes(res.locals.school) ? JSON.parse(user.teacher)[res.locals.school] : 'Pending',
                            children: !JSON.parse(user.pendingschools).includes(res.locals.school) ? fixedChildren : undefined
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

    app.put('/people', async (req: express.Request, res: express.Response) => {
        if (req.body.email && req.body.type && (req.body.type === 'administrator' || (req.body.type === 'teacher' && req.body.subject) || req.body.type === 'student')) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    if (dbRes.rows.find(x => x.id === res.locals.user).administrator.includes(res.locals.school)) {
                        const user = dbRes.rows.find(x => x.email === req.body.email);
                        if (user) {
                            let pendingschools = JSON.parse(user.pendingschools);
                            pendingschools.push(res.locals.school);
                            user.pendingschools = JSON.stringify(pendingschools);

                            if (req.body.type === 'administrator') {
                                let administrator = JSON.parse(user.administrator);
                                administrator.push(res.locals.school);
                                user.administrator = JSON.stringify(administrator);
                            }

                            if (req.body.type === 'teacher') {
                                let teacher = JSON.parse(user.teacher);
                                teacher[res.locals.school] = req.body.subject;
                                user.teacher = JSON.stringify(teacher);
                            }

                            database.query(`UPDATE users SET pendingschools = $1, administrator = $2, teacher = $3 WHERE id = $4`, [user.pendingschools, user.administrator, user.teacher, user.id], (err, dbResp) => {
                                if (!err) {
                                    dbRes.rows.forEach((userToSend: any) => {
                                        if (JSON.parse(userToSend.administrator).includes(res.locals.school)) {
                                            Array.from(websockets.get(res.locals.school)?.get(userToSend.id)?.values() ?? [])?.forEach(websocket => {
                                                websocket.send(JSON.stringify({
                                                    event: 'newUser', user: {
                                                        id: user.id,
                                                        name: user.name,
                                                        email: user.email,
                                                        subject: 'Pending',
                                                        type: 'Pending'
                                                    }
                                                }));
                                            });
                                        }
                                    });
                                    websockets?.forEach(websocket => {
                                        if (websocket.has(user.id)) {
                                            websocket.get(user.id)?.forEach(websocket => {
                                                websocket.send(JSON.stringify({ event: 'schoolInvite', schoolId: res.locals.school }));
                                            });
                                        }
                                    })
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

    app.delete('/people', async (req: express.Request, res: express.Response) => {
        if (req.body.tos) {
            database.query(`SELECT * FROM users`, async (err, dbRes) => {
                if (!err) {
                    if (dbRes.rows.find(x => x.id === res.locals.user).administrator.includes(res.locals.school)) {
                        const users = dbRes.rows.filter(x => req.body.tos.includes(x.id));

                        const cleanedUsers = users.map(x => {
                            let schools = JSON.parse(x.schools);
                            if(schools.includes(res.locals.school)) {
                                schools.splice(schools.indexOf(res.locals.school), 1);
                            }
                            x.schools = JSON.stringify(schools);

                            let pendingschools = JSON.parse(x.pendingschools);
                            if(pendingschools.includes(res.locals.school)) {
                                pendingschools.splice(pendingschools.indexOf(res.locals.school), 1);
                            }
                            x.pendingschools = JSON.stringify(pendingschools);

                            let administrator = JSON.parse(x.administrator);
                            if (administrator.includes(res.locals.school)) {
                                administrator.splice(administrator.indexOf(res.locals.school), 1);
                            }
                            x.administrator = JSON.stringify(administrator);

                            let teacher = JSON.parse(x.teacher);
                            if (teacher[res.locals.school]) {
                                delete teacher[res.locals.school];
                            }
                            x.teacher = JSON.stringify(teacher);

                            let grades = JSON.parse(x.grades);
                            if (grades[res.locals.school]) {
                                delete grades[res.locals.school];
                            }
                            x.grades = JSON.stringify(grades);

                            return x;
                        });

                        cleanedUsers.forEach(user => {
                            database.query(`UPDATE users SET pendingschools = $1, schools = $2, administrator = $3, teacher = $4, grades = $5 WHERE id = $6`, [user.pendingschools, user.schools, user.administrator, user.teacher, user.grades, user.id], (err, dbResp) => {
                                if (!err) {
                                    dbRes.rows.forEach((userToSend: any) => {
                                        if (JSON.parse(user.teacher)[res.locals.school] || JSON.parse(userToSend.administrator).includes(res.locals.school)) {
                                            Array.from(websockets.get(res.locals.school)?.get(userToSend.id)?.values() ?? [])?.forEach(websocket => {
                                                websocket.send(JSON.stringify({ event: 'deletedUser', userId: user.id }));
                                            });
                                        }
                                    });
                                } else {
                                    console.log(err);
                                    res.status(500).send({ error: "Server error." });
                                }
                            });
                        });
                        res.send({});
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