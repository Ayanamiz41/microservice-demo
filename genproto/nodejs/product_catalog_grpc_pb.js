// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// product_catalog.proto — ProductCatalogService contract (Go productcatalogservice).
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo.
//
'use strict';
var grpc = require('@grpc/grpc-js');
var product_catalog_pb = require('./product_catalog_pb.js');
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

function serialize_hipstershop_GetProductRequest(arg) {
  if (!(arg instanceof product_catalog_pb.GetProductRequest)) {
    throw new Error('Expected argument of type hipstershop.GetProductRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_GetProductRequest(buffer_arg) {
  return product_catalog_pb.GetProductRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_ListProductsResponse(arg) {
  if (!(arg instanceof product_catalog_pb.ListProductsResponse)) {
    throw new Error('Expected argument of type hipstershop.ListProductsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_ListProductsResponse(buffer_arg) {
  return product_catalog_pb.ListProductsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_Product(arg) {
  if (!(arg instanceof demo_pb.Product)) {
    throw new Error('Expected argument of type hipstershop.Product');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_Product(buffer_arg) {
  return demo_pb.Product.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_SearchProductsRequest(arg) {
  if (!(arg instanceof product_catalog_pb.SearchProductsRequest)) {
    throw new Error('Expected argument of type hipstershop.SearchProductsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_SearchProductsRequest(buffer_arg) {
  return product_catalog_pb.SearchProductsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_SearchProductsResponse(arg) {
  if (!(arg instanceof product_catalog_pb.SearchProductsResponse)) {
    throw new Error('Expected argument of type hipstershop.SearchProductsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_SearchProductsResponse(buffer_arg) {
  return product_catalog_pb.SearchProductsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


// ---------------Product Catalog----------------
//
var ProductCatalogServiceService = exports.ProductCatalogServiceService = {
  listProducts: {
    path: '/hipstershop.ProductCatalogService/ListProducts',
    requestStream: false,
    responseStream: false,
    requestType: demo_pb.Empty,
    responseType: product_catalog_pb.ListProductsResponse,
    requestSerialize: serialize_hipstershop_Empty,
    requestDeserialize: deserialize_hipstershop_Empty,
    responseSerialize: serialize_hipstershop_ListProductsResponse,
    responseDeserialize: deserialize_hipstershop_ListProductsResponse,
  },
  getProduct: {
    path: '/hipstershop.ProductCatalogService/GetProduct',
    requestStream: false,
    responseStream: false,
    requestType: product_catalog_pb.GetProductRequest,
    responseType: demo_pb.Product,
    requestSerialize: serialize_hipstershop_GetProductRequest,
    requestDeserialize: deserialize_hipstershop_GetProductRequest,
    responseSerialize: serialize_hipstershop_Product,
    responseDeserialize: deserialize_hipstershop_Product,
  },
  searchProducts: {
    path: '/hipstershop.ProductCatalogService/SearchProducts',
    requestStream: false,
    responseStream: false,
    requestType: product_catalog_pb.SearchProductsRequest,
    responseType: product_catalog_pb.SearchProductsResponse,
    requestSerialize: serialize_hipstershop_SearchProductsRequest,
    requestDeserialize: deserialize_hipstershop_SearchProductsRequest,
    responseSerialize: serialize_hipstershop_SearchProductsResponse,
    responseDeserialize: deserialize_hipstershop_SearchProductsResponse,
  },
};

exports.ProductCatalogServiceClient = grpc.makeGenericClientConstructor(ProductCatalogServiceService, 'ProductCatalogService');
