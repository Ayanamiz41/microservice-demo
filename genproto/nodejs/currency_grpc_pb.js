// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// currency.proto — CurrencyService contract (Node.js currencyservice).
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo.
//
'use strict';
var grpc = require('@grpc/grpc-js');
var currency_pb = require('./currency_pb.js');
var demo_pb = require('./demo_pb.js');

function serialize_hipstershop_CurrencyConversionRequest(arg) {
  if (!(arg instanceof currency_pb.CurrencyConversionRequest)) {
    throw new Error('Expected argument of type hipstershop.CurrencyConversionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_CurrencyConversionRequest(buffer_arg) {
  return currency_pb.CurrencyConversionRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_hipstershop_GetSupportedCurrenciesResponse(arg) {
  if (!(arg instanceof currency_pb.GetSupportedCurrenciesResponse)) {
    throw new Error('Expected argument of type hipstershop.GetSupportedCurrenciesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_GetSupportedCurrenciesResponse(buffer_arg) {
  return currency_pb.GetSupportedCurrenciesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_Money(arg) {
  if (!(arg instanceof demo_pb.Money)) {
    throw new Error('Expected argument of type hipstershop.Money');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_Money(buffer_arg) {
  return demo_pb.Money.deserializeBinary(new Uint8Array(buffer_arg));
}


// -----------------Currency service-----------------
//
var CurrencyServiceService = exports.CurrencyServiceService = {
  getSupportedCurrencies: {
    path: '/hipstershop.CurrencyService/GetSupportedCurrencies',
    requestStream: false,
    responseStream: false,
    requestType: demo_pb.Empty,
    responseType: currency_pb.GetSupportedCurrenciesResponse,
    requestSerialize: serialize_hipstershop_Empty,
    requestDeserialize: deserialize_hipstershop_Empty,
    responseSerialize: serialize_hipstershop_GetSupportedCurrenciesResponse,
    responseDeserialize: deserialize_hipstershop_GetSupportedCurrenciesResponse,
  },
  convert: {
    path: '/hipstershop.CurrencyService/Convert',
    requestStream: false,
    responseStream: false,
    requestType: currency_pb.CurrencyConversionRequest,
    responseType: demo_pb.Money,
    requestSerialize: serialize_hipstershop_CurrencyConversionRequest,
    requestDeserialize: deserialize_hipstershop_CurrencyConversionRequest,
    responseSerialize: serialize_hipstershop_Money,
    responseDeserialize: deserialize_hipstershop_Money,
  },
};

exports.CurrencyServiceClient = grpc.makeGenericClientConstructor(CurrencyServiceService, 'CurrencyService');
