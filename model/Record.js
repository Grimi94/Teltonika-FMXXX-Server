'use strict';

// Dependencies
const Mongoose = require('mongoose');
const Database = require('../db');

const RecordSchema = new Mongoose.Schema({
    imei:{
        type: String,
        unique: true,
        required: true
    },
    location:{
        type: {
            type: String,
            default: 'Point'
        },
        coordinates:{
            type: [Number],  // [<longitude>, <latitude>]
            index: '2dsphere' // create the geospatial index
        }
    },
    time:{
        type: Date
    },
    angle:{
        type: Number
    },
    speed:{
        type: Number
    }

});


module.exports = Mongoose.model('GPSRecord', RecordSchema);
