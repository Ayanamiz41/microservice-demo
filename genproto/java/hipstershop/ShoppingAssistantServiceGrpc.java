package hipstershop;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.1)",
    comments = "Source: shopping_assistant.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class ShoppingAssistantServiceGrpc {

  private ShoppingAssistantServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "hipstershop.ShoppingAssistantService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<hipstershop.ShoppingAssistant.ShoppingAssistantRequest,
      hipstershop.ShoppingAssistant.ShoppingAssistantResponse> getGetCompletionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetCompletion",
      requestType = hipstershop.ShoppingAssistant.ShoppingAssistantRequest.class,
      responseType = hipstershop.ShoppingAssistant.ShoppingAssistantResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<hipstershop.ShoppingAssistant.ShoppingAssistantRequest,
      hipstershop.ShoppingAssistant.ShoppingAssistantResponse> getGetCompletionMethod() {
    io.grpc.MethodDescriptor<hipstershop.ShoppingAssistant.ShoppingAssistantRequest, hipstershop.ShoppingAssistant.ShoppingAssistantResponse> getGetCompletionMethod;
    if ((getGetCompletionMethod = ShoppingAssistantServiceGrpc.getGetCompletionMethod) == null) {
      synchronized (ShoppingAssistantServiceGrpc.class) {
        if ((getGetCompletionMethod = ShoppingAssistantServiceGrpc.getGetCompletionMethod) == null) {
          ShoppingAssistantServiceGrpc.getGetCompletionMethod = getGetCompletionMethod =
              io.grpc.MethodDescriptor.<hipstershop.ShoppingAssistant.ShoppingAssistantRequest, hipstershop.ShoppingAssistant.ShoppingAssistantResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetCompletion"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  hipstershop.ShoppingAssistant.ShoppingAssistantRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  hipstershop.ShoppingAssistant.ShoppingAssistantResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ShoppingAssistantServiceMethodDescriptorSupplier("GetCompletion"))
              .build();
        }
      }
    }
    return getGetCompletionMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ShoppingAssistantServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ShoppingAssistantServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ShoppingAssistantServiceStub>() {
        @java.lang.Override
        public ShoppingAssistantServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ShoppingAssistantServiceStub(channel, callOptions);
        }
      };
    return ShoppingAssistantServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ShoppingAssistantServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ShoppingAssistantServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ShoppingAssistantServiceBlockingStub>() {
        @java.lang.Override
        public ShoppingAssistantServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ShoppingAssistantServiceBlockingStub(channel, callOptions);
        }
      };
    return ShoppingAssistantServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ShoppingAssistantServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ShoppingAssistantServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ShoppingAssistantServiceFutureStub>() {
        @java.lang.Override
        public ShoppingAssistantServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ShoppingAssistantServiceFutureStub(channel, callOptions);
        }
      };
    return ShoppingAssistantServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     * <pre>
     * Returns a chat-style assistant reply for the given user message.
     * </pre>
     */
    default void getCompletion(hipstershop.ShoppingAssistant.ShoppingAssistantRequest request,
        io.grpc.stub.StreamObserver<hipstershop.ShoppingAssistant.ShoppingAssistantResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetCompletionMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service ShoppingAssistantService.
   */
  public static abstract class ShoppingAssistantServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return ShoppingAssistantServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service ShoppingAssistantService.
   */
  public static final class ShoppingAssistantServiceStub
      extends io.grpc.stub.AbstractAsyncStub<ShoppingAssistantServiceStub> {
    private ShoppingAssistantServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ShoppingAssistantServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ShoppingAssistantServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * Returns a chat-style assistant reply for the given user message.
     * </pre>
     */
    public void getCompletion(hipstershop.ShoppingAssistant.ShoppingAssistantRequest request,
        io.grpc.stub.StreamObserver<hipstershop.ShoppingAssistant.ShoppingAssistantResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetCompletionMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service ShoppingAssistantService.
   */
  public static final class ShoppingAssistantServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<ShoppingAssistantServiceBlockingStub> {
    private ShoppingAssistantServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ShoppingAssistantServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ShoppingAssistantServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * Returns a chat-style assistant reply for the given user message.
     * </pre>
     */
    public hipstershop.ShoppingAssistant.ShoppingAssistantResponse getCompletion(hipstershop.ShoppingAssistant.ShoppingAssistantRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetCompletionMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service ShoppingAssistantService.
   */
  public static final class ShoppingAssistantServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<ShoppingAssistantServiceFutureStub> {
    private ShoppingAssistantServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ShoppingAssistantServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ShoppingAssistantServiceFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * Returns a chat-style assistant reply for the given user message.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<hipstershop.ShoppingAssistant.ShoppingAssistantResponse> getCompletion(
        hipstershop.ShoppingAssistant.ShoppingAssistantRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetCompletionMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_GET_COMPLETION = 0;

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
        case METHODID_GET_COMPLETION:
          serviceImpl.getCompletion((hipstershop.ShoppingAssistant.ShoppingAssistantRequest) request,
              (io.grpc.stub.StreamObserver<hipstershop.ShoppingAssistant.ShoppingAssistantResponse>) responseObserver);
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
          getGetCompletionMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              hipstershop.ShoppingAssistant.ShoppingAssistantRequest,
              hipstershop.ShoppingAssistant.ShoppingAssistantResponse>(
                service, METHODID_GET_COMPLETION)))
        .build();
  }

  private static abstract class ShoppingAssistantServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ShoppingAssistantServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return hipstershop.ShoppingAssistant.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("ShoppingAssistantService");
    }
  }

  private static final class ShoppingAssistantServiceFileDescriptorSupplier
      extends ShoppingAssistantServiceBaseDescriptorSupplier {
    ShoppingAssistantServiceFileDescriptorSupplier() {}
  }

  private static final class ShoppingAssistantServiceMethodDescriptorSupplier
      extends ShoppingAssistantServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    ShoppingAssistantServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (ShoppingAssistantServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ShoppingAssistantServiceFileDescriptorSupplier())
              .addMethod(getGetCompletionMethod())
              .build();
        }
      }
    }
    return result;
  }
}
