require('dotenv').config();

const Joi = require('joi');

const schema = Joi.object({
  // NODE_ENV: Joi.string().valid('production', 'development', 'test').required(),
  NODE_ENV: Joi.string().valid('production', 'development').required(),
  SVR_PORT: Joi.number().default(3001),
  DB_USER: Joi.string().default('postgres'),
  DB_HOST: Joi.string().default('localhost'),
  DB_PASSWORD: Joi.string().default(''),
  DB_DATABASE: Joi.string().default('postgres'),
  DB_PORT: Joi.number().default(5432),

  VAPID_PUBLIC_KEY: Joi.string().required(),
  VAPID_PRIVATE_KEY: Joi.string().required(),
  EMAIL: Joi.string().required(),
})
  .unknown();

const { error, value } = schema.prefs({ errors: { label: 'key' } }).validate(process.env);

if (error) console.error(error);

export default {
  env: value.NODE_ENV,
  server: {
    port: value.SVR_PORT,
  },
  db: {
    user: value.DB_USER,
    host: value.DB_HOST,
    password: value.DB_PASSWORD,
    database: value.DB_DATABASE,
    port: value.DB_PORT,
  },
  vapid: {
    publicKey: value.VAPID_PUBLIC_KEY,
    privateKey: value.VAPID_PRIVATE_KEY,
    email: value.EMAIL
  }
};