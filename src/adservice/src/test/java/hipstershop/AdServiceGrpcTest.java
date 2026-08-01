package hipstershop;

import static org.junit.jupiter.api.Assertions.assertEquals;

import hipstershop.AdOuterClass.AdRequest;
import hipstershop.AdOuterClass.AdResponse;
import io.grpc.ManagedChannel;
import io.grpc.Server;
import io.grpc.health.v1.HealthCheckRequest;
import io.grpc.health.v1.HealthCheckResponse.ServingStatus;
import io.grpc.health.v1.HealthGrpc;
import io.grpc.inprocess.InProcessChannelBuilder;
import io.grpc.inprocess.InProcessServerBuilder;
import java.io.IOException;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

/**
 * In-process gRPC tests for the AdService RPC and the {@code grpc.health.v1.Health} service
 * (mirrors the acceptance health check {@code {"service": "hipstershop.AdService"}} -> SERVING).
 */
class AdServiceGrpcTest {

  private static final String SERVICE_NAME = "hipstershop.AdService";

  private Server server;
  private ManagedChannel channel;
  private HealthServiceImpl healthService;

  @BeforeEach
  void setUp() throws IOException {
    healthService = new HealthServiceImpl();
    String name = InProcessServerBuilder.generateName();
    server =
        InProcessServerBuilder.forName(name)
            .directExecutor()
            .addService(new AdServiceImpl(new AdCatalog()))
            .addService(healthService)
            .build()
            .start();
    healthService.setStatus(SERVICE_NAME, ServingStatus.SERVING);
    healthService.setStatus("", ServingStatus.SERVING);
    channel = InProcessChannelBuilder.forName(name).directExecutor().build();
  }

  @AfterEach
  void tearDown() {
    channel.shutdownNow();
    server.shutdownNow();
  }

  @Test
  void getAdsReturnsMatchingAdForContextKey() {
    AdServiceGrpc.AdServiceBlockingStub stub = AdServiceGrpc.newBlockingStub(channel);
    AdResponse response =
        stub.getAds(AdRequest.newBuilder().addContextKeys("clothing").build());
    assertEquals(1, response.getAdsCount());
    assertEquals("Tank top for sale. 20% off.", response.getAds(0).getText());
  }

  @Test
  void getAdsReturnsRandomAdsForEmptyRequest() {
    AdServiceGrpc.AdServiceBlockingStub stub = AdServiceGrpc.newBlockingStub(channel);
    AdResponse response = stub.getAds(AdRequest.getDefaultInstance());
    assertEquals(AdCatalog.MAX_ADS_TO_SERVE, response.getAdsCount());
  }

  @Test
  void healthCheckReportsServingForNamedService() {
    HealthGrpc.HealthBlockingStub stub = HealthGrpc.newBlockingStub(channel);
    ServingStatus status =
        stub.check(HealthCheckRequest.newBuilder().setService(SERVICE_NAME).build()).getStatus();
    assertEquals(ServingStatus.SERVING, status);
  }

  @Test
  void healthCheckReportsServingForOverallStatus() {
    HealthGrpc.HealthBlockingStub stub = HealthGrpc.newBlockingStub(channel);
    ServingStatus status =
        stub.check(HealthCheckRequest.getDefaultInstance()).getStatus();
    assertEquals(ServingStatus.SERVING, status);
  }
}
