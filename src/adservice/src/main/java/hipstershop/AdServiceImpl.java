package hipstershop;

import hipstershop.AdOuterClass.Ad;
import hipstershop.AdOuterClass.AdRequest;
import hipstershop.AdOuterClass.AdResponse;
import hipstershop.AdServiceGrpc.AdServiceImplBase;
import io.grpc.stub.StreamObserver;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/**
 * gRPC implementation of {@code hipstershop.AdService.GetAds}.
 *
 * <p>Replicates the upstream Online Boutique adservice behavior: ads are selected from an
 * in-memory catalog by context keys, with random ads as the fallback.
 */
@Component
public class AdServiceImpl extends AdServiceImplBase {

  private static final Logger logger = LoggerFactory.getLogger(AdServiceImpl.class);

  private final AdCatalog adCatalog;

  public AdServiceImpl(AdCatalog adCatalog) {
    this.adCatalog = adCatalog;
  }

  @Override
  public void getAds(AdRequest request, StreamObserver<AdResponse> responseObserver) {
    logger.info("received ad request (context_words={})", request.getContextKeysList());
    List<Ad> ads = adCatalog.getAds(request.getContextKeysList());
    responseObserver.onNext(AdResponse.newBuilder().addAllAds(ads).build());
    responseObserver.onCompleted();
  }
}
