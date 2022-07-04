import express from "express";
import fs from 'fs';
import crypto from 'crypto';
import mime from 'mime-types';
import multer from "multer";
const upload = multer({ storage: multer.memoryStorage() })

export default (app: express.Application) => {
    app.post('/upload', upload.single('upload'), (req: express.Request, res: express.Response) => {
        const uuid = crypto.randomUUID() + '.' + mime.extension(req.file?.mimetype ?? '');
        if (req.file) {
            fs.writeFile(__dirname + '/../../files/' + uuid, req.file.buffer, "binary", (err => {
                if (!err) {
                    res.status(201).send({
                        id: uuid,
                        url: req.protocol + '://' + req.get('host') + '/static/' + uuid // The way I'm getting the URl isn't good
                    });
                } else {
                    console.log(err);
                    res.status(500).send({ error: "Server error." });
                }
            }));
        } else {
            res.status(400).send({ error: "Missing required argument." });
        }
    });
};