'use strict'

// Database configuration
const config = require('./config.json');
const Mongoose = require('mongoose');


Mongoose.connect('mongodb://' + config.database.host + '/' + config.database.db);

const db = Mongoose.connection;
db.on('error', console.error.bind(console, 'Connection error'));
db.once('open', () => {
    console.log('Connection with database succeeded');
});

module.exports = db;
