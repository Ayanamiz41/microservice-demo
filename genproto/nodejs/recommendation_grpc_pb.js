// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// recommendation.proto — RecommendationService contract (Python recommendationservice).
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo.
//
'use strict';
var grpc = require('@grpc/grpc-js');
var recommendation_pb = require('./recommendation_pb.js');

function serialize_hipstershop_ListRecommendationsRequest(arg) {
  if (!(arg instanceof recommendation_pb.ListRecommendationsRequest)) {
    throw new Error('Expected argument of type hipstershop.ListRecommendationsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_ListRecommendationsRequest(buffer_arg) {
  return recommendation_pb.ListRecommendationsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_hipstershop_ListRecommendationsResponse(arg) {
  if (!(arg instanceof recommendation_pb.ListRecommendationsResponse)) {
    throw new Error('Expected argument of type hipstershop.ListRecommendationsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_hipstershop_ListRecommendationsResponse(buffer_arg) {
  return recommendation_pb.ListRecommendationsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


// ---------------Recommendation service----------
//
var RecommendationServiceService = exports.RecommendationServiceService = {
  listRecommendations: {
    path: '/hipstershop.RecommendationService/ListRecommendations',
    requestStream: false,
    responseStream: false,
    requestType: recommendation_pb.ListRecommendationsRequest,
    responseType: recommendation_pb.ListRecommendationsResponse,
    requestSerialize: serialize_hipstershop_ListRecommendationsRequest,
    requestDeserialize: deserialize_hipstershop_ListRecommendationsRequest,
    responseSerialize: serialize_hipstershop_ListRecommendationsResponse,
    responseDeserialize: deserialize_hipstershop_ListRecommendationsResponse,
  },
};

exports.RecommendationServiceClient = grpc.makeGenericClientConstructor(RecommendationServiceService, 'RecommendationService');
