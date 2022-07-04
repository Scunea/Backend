import { Client } from "pg";

export default async (database: Client, logger: any) => {
    await database.connect();

    database.query(`CREATE TABLE IF NOT EXISTS users (
        token text NOT NULL,
        id text NOT NULL,
        name text NOT NULL,
        grades text NOT NULL,
        password text NOT NULL,
        administrator text NOT NULL,
        teacher text NOT NULL,
        parent text NOT NULL,
        schools text NOT NULL,
        PRIMARY KEY (id)
    )`, (err, dbRes) => {
        if (err) {
            logger.fatal('Something went terribly wrong initializing.');
        }
    });

    database.query(`CREATE TABLE IF NOT EXISTS schools (
        id text NOT NULL,
        name text NOT NULL,
        PRIMARY KEY (id)
    )`, (err, dbRes) => {
        if (err) {
            logger.fatal('Something went terribly wrong initializing.');
        }
    });

    database.query(`CREATE TABLE IF NOT EXISTS schedule (
        id text NOT NULL,
        title text NOT NULL,
        date text NOT NULL,
        author text NOT NULL,
        receiver text NOT NULL,
        school text NOT NULL,
        PRIMARY KEY (id)
    )`, (err, dbRes) => {
        if (err) {
            logger.fatal('Something went terribly wrong initializing.');
        }
    });

    database.query(`CREATE TABLE IF NOT EXISTS messages (
        id text NOT NULL,
        title text NOT NULL,
        content text NOT NULL,
        files text NOT NULL,
        author text NOT NULL,
        date text NOT NULL,
        receiver text NOT NULL,
        school text NOT NULL,
        PRIMARY KEY (id)
    )`, (err, dbRes) => {
        if (err) {
            logger.fatal('Something went terribly wrong initializing.');
        }
    });

    database.query(`CREATE TABLE IF NOT EXISTS reports (
        id text NOT NULL,
        title text NOT NULL,
        file text NOT NULL,
        author text NOT NULL,
        date text NOT NULL,
        school text NOT NULL,
        PRIMARY KEY (id)
    )`, (err, dbRes) => {
        if (err) {
            logger.fatal('Something went terribly wrong initializing.');
        }
    });

    database.query(`CREATE TABLE IF NOT EXISTS activities (
        id text NOT NULL,
        title text NOT NULL,
        description text NOT NULL,
        files text NOT NULL,
        type text NOT NULL,
        delivery TEXT NOT NULL,
        author text NOT NULL,
        date text NOT NULL,
        expiration text NOT NULL,
        receiver text NOT NULL,
        delivered text NOT NULL,
        result text NOT NULL,
        viewed text NOT NULL,
        school text NOT NULL,
        PRIMARY KEY (id)
    )`, (err, dbRes) => {
        if (err) {
            logger.fatal('Something went terribly wrong initializing.');
        }
    });


};