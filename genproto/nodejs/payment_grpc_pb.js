// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// payment.proto — PaymentService contract (Node.js paymentservice).
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo.
//
'use strict';
var grpc = require('@grpc/grpc-js');
var payment_pb = require('./payment_pb.js');
var demo_pb = require('./demo_pb.js');

function serialize_hipstershop_ChargeRequest(arg) {
  if (!(arg instanceof payment_pb.ChargeRequest)) {
    throw new Error('Expected argument of type hipstershop.ChargeRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_ChargeRequest(buffer_arg) {
  return payment_pb.ChargeRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_ChargeResponse(arg) {
  if (!(arg instanceof payment_pb.ChargeResponse)) {
    throw new Error('Expected argument of type hipstershop.ChargeResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_ChargeResponse(buffer_arg) {
  return payment_pb.ChargeResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


// -------------Payment service-----------------
//
var PaymentServiceService = exports.PaymentServiceService = {
  charge: {
    path: '/hipstershop.PaymentService/Charge',
    requestStream: false,
    responseStream: false,
    requestType: payment_pb.ChargeRequest,
    responseType: payment_pb.ChargeResponse,
    requestSerialize: serialize_hipstershop_ChargeRequest,
    requestDeserialize: deserialize_hipstershop_ChargeRequest,
    responseSerialize: serialize_hipstershop_ChargeResponse,
    responseDeserialize: deserialize_hipstershop_ChargeResponse,
  },
};

exports.PaymentServiceClient = grpc.makeGenericClientConstructor(PaymentServiceService, 'PaymentService');
