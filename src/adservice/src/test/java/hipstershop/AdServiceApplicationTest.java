package hipstershop;

import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.TestPropertySource;

/** Verifies the Spring Boot context loads and the embedded gRPC server starts. */
@SpringBootTest
@TestPropertySource(properties = "adservice.port=0")
class AdServiceApplicationTest {

  @Autowired private GrpcServerLifecycle grpcServerLifecycle;
  @Autowired private AdServiceImpl adServiceImpl;

  @Test
  void contextLoadsAndGrpcServerStarts() {
    assertNotNull(adServiceImpl);
    assertTrue(grpcServerLifecycle.isStarted());
  }
}
