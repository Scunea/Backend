import { Server } from 'http';
import { WebSocket, WebSocketServer } from 'ws';
import { User } from '../interfaces';
import { Client } from 'pg';
import * as jose from 'jose';
import fs from 'fs';
import crypto from 'crypto';

export default (wss: WebSocketServer, websockets: Map<string, Map<string, Map<string, WebSocket>>>, server: Server, database: Client) => {
    server.on('upgrade', async (request, socket, head) => {
        const url = new URL(request.url ?? '/', `http://${request.headers.host}`);
        const pathname = url.pathname;
        const token = url.searchParams.get('token') ?? '';
        const school = url.searchParams.get('school') ?? '';
        const user: User = await checkLogin(false, token, school);
        if (pathname === '/socket' && user.id != '') {
            wss.handleUpgrade(request, socket, head, (ws) => {
                let finalMap = websockets.get(school) ?? new Map();
                let websocketForThis = finalMap?.get(user.id) ?? new Map();
                const uuid = crypto.randomUUID();
                websocketForThis.set(uuid, ws);
                finalMap?.set(user.id, websocketForThis);
                websockets.set(school, finalMap);

                function close() {
                    socket.destroy();
                    let finalMap = websockets.get(school) ?? new Map();
                    let websocketForThis = finalMap?.get(user.id) ?? new Map();
                    websocketForThis.delete(uuid);
                    finalMap?.set(user.id, websocketForThis);
                    websockets.set(school, finalMap);
                }

                let closer = setTimeout(close, 10000);

                ws.send('Ping!');
                ws.onmessage = (event) => {
                    if(event.data === 'Pong!') {
                        clearTimeout(closer);
                        setTimeout(() => {
                            ws.send('Ping!');
                            closer = setTimeout(close, 10000);
                        }, 2000);
                    }
                }
            });
        } else {
            socket.destroy();
        }
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
}