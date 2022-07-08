import { Grade, User } from '../interfaces';

import express from "express";
import { Client } from 'pg';

export default (app: express.Application, database: Client) => {
    app.get('/info', (req: express.Request, res: express.Response) => {
        database.query(`SELECT * FROM users`, async (err, dbRes) => {
            if (!err) {
                const userRows = dbRes.rows;
                database.query(`SELECT * FROM schools`, async (err, dbRes) => {
                    if (!err) {
                        const user = { ...userRows.find(x => x.id === res.locals.user) };
                        if (user) {
                            const userPrivatelessPlusAvaliable: User = user;
                            delete userPrivatelessPlusAvaliable.token;
                            delete userPrivatelessPlusAvaliable.password;
                            delete userPrivatelessPlusAvaliable.verified;
                            delete userPrivatelessPlusAvaliable.verificator;
                            userPrivatelessPlusAvaliable.tfa = Boolean(userPrivatelessPlusAvaliable.tfa);
                            delete userPrivatelessPlusAvaliable.schools;
                            delete userPrivatelessPlusAvaliable.pendingschools;
                            delete userPrivatelessPlusAvaliable.parents;
                            delete userPrivatelessPlusAvaliable.pendingparents;
                            userPrivatelessPlusAvaliable.administrator = JSON.parse(user.administrator).includes(res.locals.school) ?? '';
                            userPrivatelessPlusAvaliable.teacher = JSON.parse(user.teacher)[res.locals.school] ?? '';
                            userPrivatelessPlusAvaliable.children = userRows.filter(x => JSON.parse(x.parents).includes(res.locals.user))?.map(x => x.id);
                            if (!userPrivatelessPlusAvaliable.administrator && !userPrivatelessPlusAvaliable.teacher && userPrivatelessPlusAvaliable.children.length < 1) {
                                const grades = JSON.parse(user.grades)[res.locals.school] ?? {};
                                userPrivatelessPlusAvaliable.grades = Object.keys(grades).map(x => {
                                    return {
                                        subject: JSON.parse(userRows.find(y => y.id === x).teacher)[res.locals.school],
                                        deliberation: grades[x].deliberation,
                                        conceptual: grades[x].conceptual,
                                        averageFirstFour: grades[x].averageFirstFour,
                                        averageSecondFour: grades[x].averageSecondFour,
                                        final: grades[x].final
                                    }
                                });
                            } else if (userPrivatelessPlusAvaliable.children.length > 0) {
                                userPrivatelessPlusAvaliable.grades = userPrivatelessPlusAvaliable.children.map(child => {
                                const grades = JSON.parse(userRows.find(y => y.id === child).grades)[res.locals.school] ?? {};
                                return Object.keys(grades).map(z => {
                                    return {
                                        id: child,
                                        fullName: userRows.find(y => y.id === child).name,
                                        subject: JSON.parse(userRows.find(y => y.id === z).teacher)[res.locals.school],
                                        deliberation: grades[z].deliberation,
                                        conceptual: grades[z].conceptual,
                                        averageFirstFour: grades[z].averageFirstFour,
                                        averageSecondFour: grades[z].averageSecondFour,
                                        final: grades[z].final
                                    }
                                });
                            }).flat();
                            } else {
                                userPrivatelessPlusAvaliable.grades = [];
                            }
                            let avaliableUsers = userRows.filter(x => JSON.parse(x.schools).includes(res.locals.school));
                            if (!JSON.parse(userRows.find(x => x.id === res.locals.user).teacher)[res.locals.school] && !JSON.parse(userRows.find(x => x.id === res.locals.user).administrator).includes(res.locals.school)) {
                                avaliableUsers = avaliableUsers.filter(x => JSON.parse(x.teacher)[res.locals.school] || JSON.parse(x.administrator).includes(res.locals.school));
                            }
                            userPrivatelessPlusAvaliable.avaliable = avaliableUsers.map(x => x.id).map(x => {
                                return {
                                    id: x,
                                    name: userRows.find(y => y.id === x).name,
                                    type: JSON.parse(userRows.find(y => y.id === x).administrator).includes(res.locals.school) ?
                                        'Administrator' : JSON.parse(userRows.find(y => y.id === x).teacher)[res.locals.school] ?
                                            'Teacher' : userRows.find(y => JSON.parse(y.schools).includes(res.locals.school) &&  JSON.parse(y.parents)?.includes(x)) ?
                                                'Parent' :
                                                'Student',
                                    teacher: JSON.parse(userRows.find(y => y.id === x).teacher)[res.locals.school],
                                    children: userRows.filter(x => JSON.parse(x.parents).includes(x))?.map(x => x.id)
                                };
                            });
                            res.send(userPrivatelessPlusAvaliable);
                        } else {
                            res.status(404).send({ error: "Not found." });
                        }
                    } else {
                        res.status(500).send({});
                    }
                });
            } else {
                console.log(err);
                res.status(500).send({ error: "Server error." });
            }
        });
    });
};