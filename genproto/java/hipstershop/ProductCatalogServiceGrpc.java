package hipstershop;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.1)",
    comments = "Source: product_catalog.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class ProductCatalogServiceGrpc {

  private ProductCatalogServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "hipstershop.ProductCatalogService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<hipstershop.Demo.Empty,
      hipstershop.ProductCatalog.ListProductsResponse> getListProductsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListProducts",
      requestType = hipstershop.Demo.Empty.class,
      responseType = hipstershop.ProductCatalog.ListProductsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<hipstershop.Demo.Empty,
      hipstershop.ProductCatalog.ListProductsResponse> getListProductsMethod() {
    io.grpc.MethodDescriptor<hipstershop.Demo.Empty, hipstershop.ProductCatalog.ListProductsResponse> getListProductsMethod;
    if ((getListProductsMethod = ProductCatalogServiceGrpc.getListProductsMethod) == null) {
      synchronized (ProductCatalogServiceGrpc.class) {
        if ((getListProductsMethod = ProductCatalogServiceGrpc.getListProductsMethod) == null) {
          ProductCatalogServiceGrpc.getListProductsMethod = getListProductsMethod =
              io.grpc.MethodDescriptor.<hipstershop.Demo.Empty, hipstershop.ProductCatalog.ListProductsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListProducts"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  hipstershop.Demo.Empty.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  hipstershop.ProductCatalog.ListProductsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProductCatalogServiceMethodDescriptorSupplier("ListProducts"))
              .build();
        }
      }
    }
    return getListProductsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<hipstershop.ProductCatalog.GetProductRequest,
      hipstershop.Demo.Product> getGetProductMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetProduct",
      requestType = hipstershop.ProductCatalog.GetProductRequest.class,
      responseType = hipstershop.Demo.Product.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<hipstershop.ProductCatalog.GetProductRequest,
      hipstershop.Demo.Product> getGetProductMethod() {
    io.grpc.MethodDescriptor<hipstershop.ProductCatalog.GetProductRequest, hipstershop.Demo.Product> getGetProductMethod;
    if ((getGetProductMethod = ProductCatalogServiceGrpc.getGetProductMethod) == null) {
      synchronized (ProductCatalogServiceGrpc.class) {
        if ((getGetProductMethod = ProductCatalogServiceGrpc.getGetProductMethod) == null) {
          ProductCatalogServiceGrpc.getGetProductMethod = getGetProductMethod =
              io.grpc.MethodDescriptor.<hipstershop.ProductCatalog.GetProductRequest, hipstershop.Demo.Product>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetProduct"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  hipstershop.ProductCatalog.GetProductRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  hipstershop.Demo.Product.getDefaultInstance()))
              .setSchemaDescriptor(new ProductCatalogServiceMethodDescriptorSupplier("GetProduct"))
              .build();
        }
      }
    }
    return getGetProductMethod;
  }

  private static volatile io.grpc.MethodDescriptor<hipstershop.ProductCatalog.SearchProductsRequest,
      hipstershop.ProductCatalog.SearchProductsResponse> getSearchProductsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SearchProducts",
      requestType = hipstershop.ProductCatalog.SearchProductsRequest.class,
      responseType = hipstershop.ProductCatalog.SearchProductsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<hipstershop.ProductCatalog.SearchProductsRequest,
      hipstershop.ProductCatalog.SearchProductsResponse> getSearchProductsMethod() {
    io.grpc.MethodDescriptor<hipstershop.ProductCatalog.SearchProductsRequest, hipstershop.ProductCatalog.SearchProductsResponse> getSearchProductsMethod;
    if ((getSearchProductsMethod = ProductCatalogServiceGrpc.getSearchProductsMethod) == null) {
      synchronized (ProductCatalogServiceGrpc.class) {
        if ((getSearchProductsMethod = ProductCatalogServiceGrpc.getSearchProductsMethod) == null) {
          ProductCatalogServiceGrpc.getSearchProductsMethod = getSearchProductsMethod =
              io.grpc.MethodDescriptor.<hipstershop.ProductCatalog.SearchProductsRequest, hipstershop.ProductCatalog.SearchProductsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SearchProducts"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  hipstershop.ProductCatalog.SearchProductsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  hipstershop.ProductCatalog.SearchProductsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProductCatalogServiceMethodDescriptorSupplier("SearchProducts"))
              .build();
        }
      }
    }
    return getSearchProductsMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ProductCatalogServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ProductCatalogServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ProductCatalogServiceStub>() {
        @java.lang.Override
        public ProductCatalogServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ProductCatalogServiceStub(channel, callOptions);
        }
      };
    return ProductCatalogServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ProductCatalogServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ProductCatalogServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ProductCatalogServiceBlockingStub>() {
        @java.lang.Override
        public ProductCatalogServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ProductCatalogServiceBlockingStub(channel, callOptions);
        }
      };
    return ProductCatalogServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ProductCatalogServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ProductCatalogServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ProductCatalogServiceFutureStub>() {
        @java.lang.Override
        public ProductCatalogServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ProductCatalogServiceFutureStub(channel, callOptions);
        }
      };
    return ProductCatalogServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void listProducts(hipstershop.Demo.Empty request,
        io.grpc.stub.StreamObserver<hipstershop.ProductCatalog.ListProductsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListProductsMethod(), responseObserver);
    }

    /**
     */
    default void getProduct(hipstershop.ProductCatalog.GetProductRequest request,
        io.grpc.stub.StreamObserver<hipstershop.Demo.Product> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetProductMethod(), responseObserver);
    }

    /**
     */
    default void searchProducts(hipstershop.ProductCatalog.SearchProductsRequest request,
        io.grpc.stub.StreamObserver<hipstershop.ProductCatalog.SearchProductsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSearchProductsMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service ProductCatalogService.
   */
  public static abstract class ProductCatalogServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return ProductCatalogServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service ProductCatalogService.
   */
  public static final class ProductCatalogServiceStub
      extends io.grpc.stub.AbstractAsyncStub<ProductCatalogServiceStub> {
    private ProductCatalogServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ProductCatalogServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ProductCatalogServiceStub(channel, callOptions);
    }

    /**
     */
    public void listProducts(hipstershop.Demo.Empty request,
        io.grpc.stub.StreamObserver<hipstershop.ProductCatalog.ListProductsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListProductsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void getProduct(hipstershop.ProductCatalog.GetProductRequest request,
        io.grpc.stub.StreamObserver<hipstershop.Demo.Product> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetProductMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void searchProducts(hipstershop.ProductCatalog.SearchProductsRequest request,
        io.grpc.stub.StreamObserver<hipstershop.ProductCatalog.SearchProductsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSearchProductsMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service ProductCatalogService.
   */
  public static final class ProductCatalogServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<ProductCatalogServiceBlockingStub> {
    private ProductCatalogServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ProductCatalogServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ProductCatalogServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public hipstershop.ProductCatalog.ListProductsResponse listProducts(hipstershop.Demo.Empty request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListProductsMethod(), getCallOptions(), request);
    }

    /**
     */
    public hipstershop.Demo.Product getProduct(hipstershop.ProductCatalog.GetProductRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetProductMethod(), getCallOptions(), request);
    }

    /**
     */
    public hipstershop.ProductCatalog.SearchProductsResponse searchProducts(hipstershop.ProductCatalog.SearchProductsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSearchProductsMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service ProductCatalogService.
   */
  public static final class ProductCatalogServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<ProductCatalogServiceFutureStub> {
    private ProductCatalogServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ProductCatalogServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ProductCatalogServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<hipstershop.ProductCatalog.ListProductsResponse> listProducts(
        hipstershop.Demo.Empty request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListProductsMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<hipstershop.Demo.Product> getProduct(
        hipstershop.ProductCatalog.GetProductRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetProductMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<hipstershop.ProductCatalog.SearchProductsResponse> searchProducts(
        hipstershop.ProductCatalog.SearchProductsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSearchProductsMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_LIST_PRODUCTS = 0;
  private static final int METHODID_GET_PRODUCT = 1;
  private static final int METHODID_SEARCH_PRODUCTS = 2;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_LIST_PRODUCTS:
          serviceImpl.listProducts((hipstershop.Demo.Empty) request,
              (io.grpc.stub.StreamObserver<hipstershop.ProductCatalog.ListProductsResponse>) responseObserver);
          break;
        case METHODID_GET_PRODUCT:
          serviceImpl.getProduct((hipstershop.ProductCatalog.GetProductRequest) request,
              (io.grpc.stub.StreamObserver<hipstershop.Demo.Product>) responseObserver);
          break;
        case METHODID_SEARCH_PRODUCTS:
          serviceImpl.searchProducts((hipstershop.ProductCatalog.SearchProductsRequest) request,
              (io.grpc.stub.StreamObserver<hipstershop.ProductCatalog.SearchProductsResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getListProductsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              hipstershop.Demo.Empty,
              hipstershop.ProductCatalog.ListProductsResponse>(
                service, METHODID_LIST_PRODUCTS)))
        .addMethod(
          getGetProductMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              hipstershop.ProductCatalog.GetProductRequest,
              hipstershop.Demo.Product>(
                service, METHODID_GET_PRODUCT)))
        .addMethod(
          getSearchProductsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              hipstershop.ProductCatalog.SearchProductsRequest,
              hipstershop.ProductCatalog.SearchProductsResponse>(
                service, METHODID_SEARCH_PRODUCTS)))
        .build();
  }

  private static abstract class ProductCatalogServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ProductCatalogServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return hipstershop.ProductCatalog.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("ProductCatalogService");
    }
  }

  private static final class ProductCatalogServiceFileDescriptorSupplier
      extends ProductCatalogServiceBaseDescriptorSupplier {
    ProductCatalogServiceFileDescriptorSupplier() {}
  }

  private static final class ProductCatalogServiceMethodDescriptorSupplier
      extends ProductCatalogServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    ProductCatalogServiceMethodDescriptorSupplier(java.lang.String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (ProductCatalogServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ProductCatalogServiceFileDescriptorSupplier())
              .addMethod(getListProductsMethod())
              .addMethod(getGetProductMethod())
              .addMethod(getSearchProductsMethod())
              .build();
        }
      }
    }
    return result;
  }
}
