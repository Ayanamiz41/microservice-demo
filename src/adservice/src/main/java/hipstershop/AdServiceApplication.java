package hipstershop;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;

/**
 * Spring Boot entry point for the AdService (Online Boutique adservice).
 *
 * <p>Starts the embedded gRPC server (see {@link GrpcServerLifecycle}) serving the
 * {@code hipstershop.AdService} RPC plus the standard {@code grpc.health.v1.Health} service.
 */
@SpringBootApplication
public class AdServiceApplication {

  public static void main(String[] args) throws Exception {
    ConfigurableApplicationContext context =
        SpringApplication.run(AdServiceApplication.class, args);
    // grpc-java runs its server on daemon threads; block the main thread so the
    // process stays alive until the server is shut down (mirrors upstream).
    context.getBean(GrpcServerLifecycle.class).awaitTermination();
  }
}
