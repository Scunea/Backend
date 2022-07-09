import express from "express";
import fs from 'fs';
import crypto from 'crypto';
import mime from 'mime-types';
import multer from "multer";
const upload = multer({ storage: multer.memoryStorage() })

export default (app: express.Application) => {
    app.post('/upload', upload.array('upload'), (req: express.Request, res: express.Response) => {
        let uuids: string[] = [];
        if (Array.isArray(req.files)) {
            for (const file of req.files) {
                const uuid = crypto.randomUUID() + '.' + mime.extension(file?.mimetype ?? '');
                uuids.push(uuid);
                fs.writeFileSync(__dirname + '/../../files/' + uuid, file.buffer, "binary");
            }
            const preparedResponse = uuids.map(uuid => {
                return {
                    id: uuid,
                    url: req.protocol + '://' + req.get('host') + '/static/' + uuid
                }
            });
            res.status(201).send(!req.headers.simple ? preparedResponse : preparedResponse[0]);
        } else {
            res.status(400).send({ error: "Missing required argument." });
        }
    });
};