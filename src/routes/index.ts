import express from 'express';
import { Client } from 'pg';
import * as jose from 'jose';
import fs from 'fs';
import { User } from '../interfaces';
import morgan from 'morgan';
import { transliterate } from 'transliteration';

import login from './login';
import info from './info';
import grades from './grades';
import messages from './messages';
import reports from './reports';
import account from './account';
import parents from './parents';
import school from './school';
import people from './people';
import activities from './activities';
import upload from './upload';

export default (app: express.Application, database: Client, websockets: Map<string, Map<string, Map<string, WebSocket>>>) => {
    app.use(morgan('combined'));

    app.use('/static', express.static(__dirname + '/../../files', {
        setHeaders: (res) => {
            res.set('Content-Security-Policy', `Frame-Ancestors 'self' *`);
            res.set('Cross-Origin-Resource-Policy', `cross-origin`);
            if (res.req.query.name) {
                res.set('Content-Disposition', `attachment; filename="${transliterate(res.req.query.name.toString())}"`);
            }
        }
    }));

    login(app, database, checkLogin);

    app.use(async (req: express.Request, res: express.Response, next: express.NextFunction) => {
        const user: User = await checkLogin(req.url.startsWith('/create') || req.url.startsWith('/join'), req.headers.authorization ?? '', typeof req.headers.school === 'string' ? req.headers.school ?? '' : '');
        if (user.id !== '') {
            res.locals.user = user.id;
            res.locals.school = req.headers.school;
            next();
        } else {
            res.status(403).send({ error: "Not authorized." });
        }
    });

    info(app, database);

    grades(app, database, websockets);

    messages(app, database, websockets);

    reports(app, database, websockets);

    account(app, database, websockets);

    parents(app, database, websockets);

    school(app, database, websockets);

    people(app, database, websockets);

    activities(app, database, websockets);

    upload(app);

    app.use((req: express.Request, res: express.Response, next: express.NextFunction) => {
        res.status(404).send({ error: "Not found." });
    });

    app.use((err: any, req: express.Request, res: express.Response, next: express.NextFunction) => {
        console.log(err);
        console.log(err);
        res.status(500).send({ error: "Server error." });
    });

    async function checkLogin(isLogin: boolean, token: string, school?: string): Promise<User> {
        return await new Promise(resolve => {
            const emptyUser: User = {
                token: '',
                id: '',
                name: '',
                email: '',
                password: '',
                tfa: '',
                administrator: [],
                teacher: '',
                parents: [],
                grades: [],
                avaliable: [],
                schools: []
            };
            database.query(`SELECT * FROM users`, async (err, res) => {
                if (!err) {
                    const tokenExists = res.rows.map(x => x.token === token).includes(true);
                    let schoolExists = true;
                    if (tokenExists && !isLogin) {
                        schoolExists = JSON.parse(res.rows.find(x => x.token === token).schools).includes(school);
                    }
                    if (tokenExists && schoolExists) {
                        try {
                            const ecPublicKey = await jose.importSPKI(fs.readFileSync(__dirname + '/../../public.key').toString(), 'ES256');

                            const info = await jose.jwtVerify(token.split('Bearer ')[1], ecPublicKey, {
                                issuer: 'school',
                                audience: 'school'
                            });
                            resolve(res.rows.find(x => x.token === token));

                        } catch {
                            resolve(emptyUser);
                        }
                    } else {
                        resolve(emptyUser);
                    }
                } else {
                    resolve(emptyUser);
                }
            });
        });
    }
};