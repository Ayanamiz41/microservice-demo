// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// cart.proto — CartService contract (C# cartservice / Redis).
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo.
//
'use strict';
var grpc = require('@grpc/grpc-js');
var cart_pb = require('./cart_pb.js');
var demo_pb = require('./demo_pb.js');

function serialize_hipstershop_AddItemRequest(arg) {
  if (!(arg instanceof cart_pb.AddItemRequest)) {
    throw new Error('Expected argument of type hipstershop.AddItemRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_AddItemRequest(buffer_arg) {
  return cart_pb.AddItemRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_Cart(arg) {
  if (!(arg instanceof demo_pb.Cart)) {
    throw new Error('Expected argument of type hipstershop.Cart');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_Cart(buffer_arg) {
  return demo_pb.Cart.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_Empty(arg) {
  if (!(arg instanceof demo_pb.Empty)) {
    throw new Error('Expected argument of type hipstershop.Empty');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_Empty(buffer_arg) {
  return demo_pb.Empty.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_EmptyCartRequest(arg) {
  if (!(arg instanceof cart_pb.EmptyCartRequest)) {
    throw new Error('Expected argument of type hipstershop.EmptyCartRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_EmptyCartRequest(buffer_arg) {
  return cart_pb.EmptyCartRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_GetCartRequest(arg) {
  if (!(arg instanceof cart_pb.GetCartRequest)) {
    throw new Error('Expected argument of type hipstershop.GetCartRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_GetCartRequest(buffer_arg) {
  return cart_pb.GetCartRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


// -----------------Cart service-----------------
//
var CartServiceService = exports.CartServiceService = {
  addItem: {
    path: '/hipstershop.CartService/AddItem',
    requestStream: false,
    responseStream: false,
    requestType: cart_pb.AddItemRequest,
    responseType: demo_pb.Empty,
    requestSerialize: serialize_hipstershop_AddItemRequest,
    requestDeserialize: deserialize_hipstershop_AddItemRequest,
    responseSerialize: serialize_hipstershop_Empty,
    responseDeserialize: deserialize_hipstershop_Empty,
  },
  getCart: {
    path: '/hipstershop.CartService/GetCart',
    requestStream: false,
    responseStream: false,
    requestType: cart_pb.GetCartRequest,
    responseType: demo_pb.Cart,
    requestSerialize: serialize_hipstershop_GetCartRequest,
    requestDeserialize: deserialize_hipstershop_GetCartRequest,
    responseSerialize: serialize_hipstershop_Cart,
    responseDeserialize: deserialize_hipstershop_Cart,
  },
  emptyCart: {
    path: '/hipstershop.CartService/EmptyCart',
    requestStream: false,
    responseStream: false,
    requestType: cart_pb.EmptyCartRequest,
    responseType: demo_pb.Empty,
    requestSerialize: serialize_hipstershop_EmptyCartRequest,
    requestDeserialize: deserialize_hipstershop_EmptyCartRequest,
    responseSerialize: serialize_hipstershop_Empty,
    responseDeserialize: deserialize_hipstershop_Empty,
  },
};

exports.CartServiceClient = grpc.makeGenericClientConstructor(CartServiceService, 'CartService');
