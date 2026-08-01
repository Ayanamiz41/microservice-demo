// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// shopping_assistant.proto — ShoppingAssistantService contract (Python
// shoppingassistantservice).
//
// NOTE: upstream GoogleCloudPlatform/microservices-demo implements this
// service as a cloud-only Flask app backed by AlloyDB/Gemini and defines no
// gRPC contract. This replica replaces it with a simple local gRPC assistant
// (rule/catalog based); the contract below is the simplified in-repo
// definition for that scope.
//
'use strict';
var grpc = require('@grpc/grpc-js');
var shopping_assistant_pb = require('./shopping_assistant_pb.js');

function serialize_hipstershop_ShoppingAssistantRequest(arg) {
  if (!(arg instanceof shopping_assistant_pb.ShoppingAssistantRequest)) {
    throw new Error('Expected argument of type hipstershop.ShoppingAssistantRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_ShoppingAssistantRequest(buffer_arg) {
  return shopping_assistant_pb.ShoppingAssistantRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_ShoppingAssistantResponse(arg) {
  if (!(arg instanceof shopping_assistant_pb.ShoppingAssistantResponse)) {
    throw new Error('Expected argument of type hipstershop.ShoppingAssistantResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_ShoppingAssistantResponse(buffer_arg) {
  return shopping_assistant_pb.ShoppingAssistantResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


// -----------------Shopping assistant service-----------------
//
var ShoppingAssistantServiceService = exports.ShoppingAssistantServiceService = {
  // Returns a chat-style assistant reply for the given user message.
getCompletion: {
    path: '/hipstershop.ShoppingAssistantService/GetCompletion',
    requestStream: false,
    responseStream: false,
    requestType: shopping_assistant_pb.ShoppingAssistantRequest,
    responseType: shopping_assistant_pb.ShoppingAssistantResponse,
    requestSerialize: serialize_hipstershop_ShoppingAssistantRequest,
    requestDeserialize: deserialize_hipstershop_ShoppingAssistantRequest,
    responseSerialize: serialize_hipstershop_ShoppingAssistantResponse,
    responseDeserialize: deserialize_hipstershop_ShoppingAssistantResponse,
  },
};

exports.ShoppingAssistantServiceClient = grpc.makeGenericClientConstructor(ShoppingAssistantServiceService, 'ShoppingAssistantService');
