package hipstershop;

import io.grpc.Status;
import io.grpc.health.v1.HealthCheckRequest;
import io.grpc.health.v1.HealthCheckResponse;
import io.grpc.health.v1.HealthCheckResponse.ServingStatus;
import io.grpc.health.v1.HealthGrpc;
import io.grpc.stub.StreamObserver;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import org.springframework.stereotype.Component;

/**
 * Implementation of the standard {@code grpc.health.v1.Health} service, built against the
 * committed {@code protos/grpc/health/v1/health.proto} contract.
 *
 * <p>Registered services report {@link ServingStatus#SERVING}; an unregistered service name is
 * answered with {@code NOT_FOUND} (per the gRPC health-checking spec).
 */
@Component
public class HealthServiceImpl extends HealthGrpc.HealthImplBase {

  private final Map<String, ServingStatus> statuses = new ConcurrentHashMap<>();

  public void setStatus(String service, ServingStatus status) {
    statuses.put(service, status);
  }

  public void clearStatus(String service) {
    statuses.remove(service);
  }

  @Override
  public void check(
      HealthCheckRequest request, StreamObserver<HealthCheckResponse> responseObserver) {
    ServingStatus status = statuses.get(request.getService());
    if (status == null) {
      responseObserver.onError(
          Status.NOT_FOUND
              .withDescription("unknown service " + request.getService())
              .asRuntimeException());
      return;
    }
    responseObserver.onNext(HealthCheckResponse.newBuilder().setStatus(status).build());
    responseObserver.onCompleted();
  }
}
