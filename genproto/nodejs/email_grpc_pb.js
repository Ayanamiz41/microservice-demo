// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// email.proto — EmailService contract (Python emailservice).
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo.
//
'use strict';
var grpc = require('@grpc/grpc-js');
var email_pb = require('./email_pb.js');
var demo_pb = require('./demo_pb.js');

function serialize_hipstershop_Empty(arg) {
  if (!(arg instanceof demo_pb.Empty)) {
    throw new Error('Expected argument of type hipstershop.Empty');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_Empty(buffer_arg) {
  return demo_pb.Empty.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_SendOrderConfirmationRequest(arg) {
  if (!(arg instanceof email_pb.SendOrderConfirmationRequest)) {
    throw new Error('Expected argument of type hipstershop.SendOrderConfirmationRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_SendOrderConfirmationRequest(buffer_arg) {
  return email_pb.SendOrderConfirmationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


// -------------Email service-----------------
//
var EmailServiceService = exports.EmailServiceService = {
  sendOrderConfirmation: {
    path: '/hipstershop.EmailService/SendOrderConfirmation',
    requestStream: false,
    responseStream: false,
    requestType: email_pb.SendOrderConfirmationRequest,
    responseType: demo_pb.Empty,
    requestSerialize: serialize_hipstershop_SendOrderConfirmationRequest,
    requestDeserialize: deserialize_hipstershop_SendOrderConfirmationRequest,
    responseSerialize: serialize_hipstershop_Empty,
    responseDeserialize: deserialize_hipstershop_Empty,
  },
};

exports.EmailServiceClient = grpc.makeGenericClientConstructor(EmailServiceService, 'EmailService');
