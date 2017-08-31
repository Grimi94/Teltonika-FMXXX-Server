'use strict';

// Dependencies
const Mongoose = require('mongoose');
const Database = require('../db');

const PlaceSchema = new Mongoose.Schema({
    internalid: {
        type: String,
        unique: true,
        required: true
    },
    location: {
        type: {
            type: String,
            default: 'Point'
        },
        coordinates:{
            type: [Number],  // [<longitude>, <latitude>]
            index: '2dsphere' // create the geospatial index
        }
    },
    radius: {
        type: Number,
        required: true
    }
});

module.exports = Mongoose.model('Places', PlaceSchema);
