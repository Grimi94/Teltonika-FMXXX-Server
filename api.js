'use strict';

const Hapi = require('hapi');
const joi = require ('joi');
const config = require('./config.json');
const Place = require('./model/Place');
const Record = require('./model/Record');

const server = new Hapi.Server();
const EARTH_RADIUS = 6378.1; // This is in kilometers

server.connection({ port: config.server.port, host: config.server.host });

function addDays(date, days) {
    var result = new Date(date);
    result.setDate(result.getDate() + days);
    return result;
}

server.route({
    method: 'GET',
    path: '/validator/{imei}',
    config: {
        validate: {
            params:{
                imei: joi.string().length(34)
            },
            query: {
                internalid: joi.number().integer(), // Location internal id
                time: joi.date() // Time in UTC
            }
        }
    },
    handler: async function (request, reply) {
        const start = new Date(request.query.time);
        const end   = addDays(start, 1);

        try{
            const place = await Place.findOne({ internalid: request.query.internalid}).exec();
            if(!place){
                return reply("Unable to validate: Undefined location");
            }

            const area = {
                center: place.location.coordinates,
                radius: (place.radius / 1000.0) / EARTH_RADIUS, // Convert radius to km
                unique: true,
                spherical: true
            };

            console.log("Validating:", request.params.imei,
                        "With coordinates:", area.center,
                        "And Radius:", area.radius);
            console.log("Date From:", start, "To:", end);

            const records = await Record.find().
                where('time').gte(start).lt(end).
                where('location').within().circle(area).
                sort({ time: 'asc' }).exec();

            return reply(records);
        } catch (err) {
            return reply(err);
        }
    }
});

server.route({
    method: 'GET',
    path: '/place/{internalid}',
    config: {
        validate: {
            params:{
                internalid: joi.number().integer()
            }
        }
    },
    handler: async function(request, reply) {
        console.log("Retrieving information");
        var response;

        try {
            const place = await Place.findOne({internalid: request.params.internalid}).exec();
            return reply(place);
        } catch (err) {
            return reply(err);
        }
    }

});

server.route({
    method: 'POST',
    path: '/place/{internalid}',
    config: {
        validate: {
            params: {
                internalid: joi.number().integer()
            },
            payload: {
                latitude: joi.number().min(-90).max(90),
                longitude: joi.number().min(-180).max(180),
                radius: joi.number().integer().min(0) // Radius is in meters
            }
        }
    },
    handler: async function(request, reply) {
        console.log('Inserting New Place:', request.params.internalid);
        console.log("Latitude", request.payload.latitude);
        console.log("Longitude", request.payload.longitude);

        try {
            const place = new Place({
                internalid: request.params.internalid,
                location: {
                    type: "Point",
                    coordinates: [request.payload.longitude, request.payload.latitude]
                },
                radius: request.payload.radius
            });

            const savedPlace = await place.save();
            return reply(savedPlace);
        } catch(err) {
            return reply(err);
        }
    }
});

server.route({
    method: 'PUT',
    path: '/place/{internalid}',
    config: {
        validate: {
            params: {
                internalid: joi.number()
            },
            payload: {
                latitude: joi.number().min(-90).max(90),
                longitude: joi.number().min(-180).max(180),
                radius: joi.number().min(0)
            }
        }
    },
    handler: async function(request, reply) {
        try {
            const updated = await Place.update({
                internalid: request.params.internalid
            }, {
                radius: request.payload.radius,
                location: {
                    type: "Point",
                    coordinates: [request.payload.longitude, request.payload.latitude]
                }
            }, {upsert: true}).exec();
            return reply(updated);
        } catch (err){
            return reply(err);
        }
    }
});

server.route({
    method: 'DELETE',
    path: '/place/{internalid}',
    config:{
        validate: {
            params:{
                internalid: joi.number()
            }
        }
    },
    handler: async function(request, reply) {
        try {
            const place = await Place.remove({internalid: request.params.internalid}).exec();
            return reply(place);
        } catch(err) {
            return reply(err);
        }
    }
});

server.start((err) => {
    if (err) {
        throw err;
    }
    console.log(`Server running at: ${server.info.uri}`);
});
