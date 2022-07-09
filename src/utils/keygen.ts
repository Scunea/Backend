import * as jose from 'jose';
import fs from 'fs';

export default async (logger: any) => {
    if (!fs.existsSync(__dirname + '/../../private.key') && !fs.existsSync(__dirname + '/../../public.key')) {
        const { publicKey, privateKey } = await jose.generateKeyPair('ES256');
        const pkcs8Pem = await jose.exportPKCS8(privateKey);
        const spkiPem = await jose.exportSPKI(publicKey);
        fs.writeFileSync(__dirname + '/../../private.key', pkcs8Pem);
        fs.writeFileSync(__dirname + '/../../public.key', spkiPem);
        logger.info('JWT keys generated');
    }
};