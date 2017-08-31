'use strict';

const Hapi = require('hapi');
const joi = require ('joi');
const config = require('./config.json');
const Place = require('./model/Place');
const Record = require('./model/Record');

const server = new Hapi.Server();
const EARTH_RADIUS = 6378.1; // This is kilometers

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
                imei: joi.string()
            },
            query: {
                internalid: joi.number(), // Location internal id
                radius: joi.number().min(0),
                time: joi.date()
            }
        }
    },
    handler: function (request, reply) {
        const start = new Date(request.query.time);
        const end   = addDays(start, 1);

        Place.findOne({ internalid: request.query.internalid}).
            exec(function(err, place){
                if(err){
                    reply("Unable to validate: Undefined location");
                    return;
                }

                const area = {
                    center: place.location.coordinates,
                    radius: request.query.radius / EARTH_RADIUS,
                    unique: true,
                    spherical: true
                };

                console.log("Validating:", request.params.imei, "With coordinates:", area.center, "And Radius:", area.radius);
                console.log("Date From:", start, "To:", end);

                Record.find().
                    where('time').gte(start).lt(end).
                    where('location').within().circle(area).
                    sort({ time: 'asc' }).
                    limit(10).
                    exec(function(err, record){
                        if(err){
                            reply(err);
                        } else {
                            reply(record);
                        }
                    });

            });
    }
});

server.route({
    method: 'GET',
    path: '/place/{internalid}',
    config: {
        validate: {
            params:{
                internalid: joi.number()
            }
        }
    },
    handler: function(request, reply) {
        console.log("Retrieving information");
        var response;

        Place.findOne({internalid: request.params.internalid})
            .exec(function(err, place){
                if(err){
                    response = err;
                } else {
                    reponse = place;
                }

            });

        reply(response);
    }

});

server.route({
    method: 'POST',
    path: '/place/{internalid}',
    config: {
        validate: {
            params: {
                internalid: joi.number()
            },
            payload: {
                latitude: joi.number(),
                longitude: joi.number(),
                radius: joi.number().integer().min(0)
            }
        }
    },
    handler: function(request, reply) {
        console.log('Inserting New Place:', request.params.internalid);
        console.log("Latitude", request.payload.latitude);
        console.log("Longitude", request.payload.longitude);

        var response;
        var place = new Place({
            internalid: request.params.internalid,
            location: {
                type: "Point",
                coordinates: [request.payload.longitude, request.payload.latitude]
            },
            radius: request.payload.radius
        });

        place.save(function(err){
            if(err){
                console.log(err);
                response = err;
            } else {
                response = "Succesful";
            }
        });

        reply(response);
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
                latitude: joi.number(),
                longitude: joi.number(),
                radius: joi.number()
            }
        }
    },
    handler: function(request, reply) {
        Place.update({
            internalid: request.params.internalid
        },{
            location: [request.payload.longitude, request.payload.latitude]
        }, {upsert: true}, function(err){
            if(err){
                reply(err);
            } else {
                reply("Succesful");
            }
        });
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
    handler: function(request, reply) {

    }
});

server.start((err) => {
    if (err) {
        throw err;
    }
    console.log(`Server running at: ${server.info.uri}`);
});
