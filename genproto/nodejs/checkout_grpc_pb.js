// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// checkout.proto — CheckoutService contract (Go checkoutservice).
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo.
//
'use strict';
var grpc = require('@grpc/grpc-js');
var checkout_pb = require('./checkout_pb.js');
var demo_pb = require('./demo_pb.js');

function serialize_hipstershop_PlaceOrderRequest(arg) {
  if (!(arg instanceof checkout_pb.PlaceOrderRequest)) {
    throw new Error('Expected argument of type hipstershop.PlaceOrderRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_PlaceOrderRequest(buffer_arg) {
  return checkout_pb.PlaceOrderRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_PlaceOrderResponse(arg) {
  if (!(arg instanceof checkout_pb.PlaceOrderResponse)) {
    throw new Error('Expected argument of type hipstershop.PlaceOrderResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_PlaceOrderResponse(buffer_arg) {
  return checkout_pb.PlaceOrderResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


// -----------------Checkout service------------------
//
var CheckoutServiceService = exports.CheckoutServiceService = {
  placeOrder: {
    path: '/hipstershop.CheckoutService/PlaceOrder',
    requestStream: false,
    responseStream: false,
    requestType: checkout_pb.PlaceOrderRequest,
    responseType: checkout_pb.PlaceOrderResponse,
    requestSerialize: serialize_hipstershop_PlaceOrderRequest,
    requestDeserialize: deserialize_hipstershop_PlaceOrderRequest,
    responseSerialize: serialize_hipstershop_PlaceOrderResponse,
    responseDeserialize: deserialize_hipstershop_PlaceOrderResponse,
  },
};

exports.CheckoutServiceClient = grpc.makeGenericClientConstructor(CheckoutServiceService, 'CheckoutService');
