package hipstershop;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.health.v1.HealthCheckResponse.ServingStatus;
import io.grpc.protobuf.services.ProtoReflectionService;
import jakarta.annotation.PreDestroy;
import java.io.IOException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

/**
 * Owns the gRPC server lifecycle inside the Spring Boot application.
 *
 * <p>Serves {@code hipstershop.AdService} plus the standard {@code grpc.health.v1.Health} service
 * (see {@link HealthServiceImpl}), so the process can be health-checked with:
 *
 * <pre>grpcurl -plaintext -d '{"service": "hipstershop.AdService"}' localhost:9555 grpc.health.v1.Health/Check</pre>
 */
@Component
public class GrpcServerLifecycle {

  private static final Logger logger = LoggerFactory.getLogger(GrpcServerLifecycle.class);

  private final Server server;
  private final HealthServiceImpl healthService;

  public GrpcServerLifecycle(
      AdServiceImpl adService,
      HealthServiceImpl healthService,
      @Value("${adservice.port:9555}") int port,
      @Value("${adservice.health.service-name:hipstershop.AdService}") String serviceName)
      throws IOException {
    this.healthService = healthService;
    server =
        ServerBuilder.forPort(port)
            .addService(adService)
            .addService(healthService)
            .addService(ProtoReflectionService.newInstance())
            .build()
            .start();
    healthService.setStatus(serviceName, ServingStatus.SERVING);
    healthService.setStatus("", ServingStatus.SERVING);
    logger.info("Ad Service started, listening on {}", port);
  }

  public boolean isStarted() {
    return server != null && !server.isShutdown();
  }

  /**
   * Blocks until the gRPC server terminates. Called from {@code main} to keep the JVM alive,
   * mirroring upstream's {@code blockUntilShutdown()}: grpc-java runs its server on daemon
   * threads, so without this the Spring Boot main thread would exit and the process would stop.
   */
  public void awaitTermination() throws InterruptedException {
    if (server != null) {
      server.awaitTermination();
    }
  }

  @PreDestroy
  public void stop() {
    if (server != null) {
      healthService.clearStatus("");
      server.shutdownNow();
      logger.info("Ad Service stopped.");
    }
  }
}
