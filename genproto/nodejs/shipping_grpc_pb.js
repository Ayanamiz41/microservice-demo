// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// shipping.proto — ShippingService contract (Go shippingservice).
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo.
//
'use strict';
var grpc = require('@grpc/grpc-js');
var shipping_pb = require('./shipping_pb.js');
var demo_pb = require('./demo_pb.js');

function serialize_hipstershop_GetQuoteRequest(arg) {
  if (!(arg instanceof shipping_pb.GetQuoteRequest)) {
    throw new Error('Expected argument of type hipstershop.GetQuoteRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_GetQuoteRequest(buffer_arg) {
  return shipping_pb.GetQuoteRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_GetQuoteResponse(arg) {
  if (!(arg instanceof shipping_pb.GetQuoteResponse)) {
    throw new Error('Expected argument of type hipstershop.GetQuoteResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_GetQuoteResponse(buffer_arg) {
  return shipping_pb.GetQuoteResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_ShipOrderRequest(arg) {
  if (!(arg instanceof shipping_pb.ShipOrderRequest)) {
    throw new Error('Expected argument of type hipstershop.ShipOrderRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_ShipOrderRequest(buffer_arg) {
  return shipping_pb.ShipOrderRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_ShipOrderResponse(arg) {
  if (!(arg instanceof shipping_pb.ShipOrderResponse)) {
    throw new Error('Expected argument of type hipstershop.ShipOrderResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_ShipOrderResponse(buffer_arg) {
  return shipping_pb.ShipOrderResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


// ---------------Shipping Service----------
//
var ShippingServiceService = exports.ShippingServiceService = {
  getQuote: {
    path: '/hipstershop.ShippingService/GetQuote',
    requestStream: false,
    responseStream: false,
    requestType: shipping_pb.GetQuoteRequest,
    responseType: shipping_pb.GetQuoteResponse,
    requestSerialize: serialize_hipstershop_GetQuoteRequest,
    requestDeserialize: deserialize_hipstershop_GetQuoteRequest,
    responseSerialize: serialize_hipstershop_GetQuoteResponse,
    responseDeserialize: deserialize_hipstershop_GetQuoteResponse,
  },
  shipOrder: {
    path: '/hipstershop.ShippingService/ShipOrder',
    requestStream: false,
    responseStream: false,
    requestType: shipping_pb.ShipOrderRequest,
    responseType: shipping_pb.ShipOrderResponse,
    requestSerialize: serialize_hipstershop_ShipOrderRequest,
    requestDeserialize: deserialize_hipstershop_ShipOrderRequest,
    responseSerialize: serialize_hipstershop_ShipOrderResponse,
    responseDeserialize: deserialize_hipstershop_ShipOrderResponse,
  },
};

exports.ShippingServiceClient = grpc.makeGenericClientConstructor(ShippingServiceService, 'ShippingService');
