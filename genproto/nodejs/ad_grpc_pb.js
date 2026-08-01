// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// ad.proto — AdService contract (Java adservice / Spring Boot + Gradle).
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo.
//
'use strict';
var grpc = require('@grpc/grpc-js');
var ad_pb = require('./ad_pb.js');

function serialize_hipstershop_AdRequest(arg) {
  if (!(arg instanceof ad_pb.AdRequest)) {
    throw new Error('Expected argument of type hipstershop.AdRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_AdRequest(buffer_arg) {
  return ad_pb.AdRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_AdResponse(arg) {
  if (!(arg instanceof ad_pb.AdResponse)) {
    throw new Error('Expected argument of type hipstershop.AdResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_AdResponse(buffer_arg) {
  return ad_pb.AdResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


// ------------Ad service------------------
//
var AdServiceService = exports.AdServiceService = {
  getAds: {
    path: '/hipstershop.AdService/GetAds',
    requestStream: false,
    responseStream: false,
    requestType: ad_pb.AdRequest,
    responseType: ad_pb.AdResponse,
    requestSerialize: serialize_hipstershop_AdRequest,
    requestDeserialize: deserialize_hipstershop_AdRequest,
    responseSerialize: serialize_hipstershop_AdResponse,
    responseDeserialize: deserialize_hipstershop_AdResponse,
  },
};

exports.AdServiceClient = grpc.makeGenericClientConstructor(AdServiceService, 'AdService');
