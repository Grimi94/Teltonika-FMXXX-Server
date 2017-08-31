'use strict';

const Hapi = require('hapi');
const joi = require ('joi');
const config = require('./config.json');
const Place = require('./model/Place');
const Record = require('./model/Record');

const server = new Hapi.Server();

server.connection({ port: config.server.port, host: config.server.host });

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
