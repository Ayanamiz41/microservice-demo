package hipstershop;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;

import hipstershop.AdOuterClass.Ad;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Random;
import org.junit.jupiter.api.Test;

/** Unit tests for the in-memory ad catalog / selection logic. */
class AdCatalogTest {

  private static final String TANK_TOP_TEXT = "Tank top for sale. 20% off.";
  private static final String MUG_TEXT = "Mug for sale. Buy two, get third one for free";

  @Test
  void returnsAdsForMatchingContextKey() {
    AdCatalog catalog = new AdCatalog(new Random(1));
    List<Ad> ads = catalog.getAds(Collections.singletonList("clothing"));
    assertEquals(1, ads.size());
    assertEquals(TANK_TOP_TEXT, ads.get(0).getText());
  }

  @Test
  void returnsUnionForMultipleContextKeys() {
    AdCatalog catalog = new AdCatalog(new Random(1));
    List<Ad> ads = catalog.getAds(Arrays.asList("clothing", "kitchen"));
    assertEquals(3, ads.size()); // 1 clothing ad + 2 kitchen ads
    assertEquals(TANK_TOP_TEXT, ads.get(0).getText());
    assertEquals(MUG_TEXT, ads.get(2).getText());
  }

  @Test
  void fallsBackToRandomAdsForUnknownContextKey() {
    AdCatalog catalog = new AdCatalog(new Random(7));
    List<Ad> ads = catalog.getAds(Collections.singletonList("unknown-category"));
    assertEquals(AdCatalog.MAX_ADS_TO_SERVE, ads.size());
  }

  @Test
  void fallsBackToRandomAdsForEmptyContextKeys() {
    AdCatalog catalog = new AdCatalog(new Random(7));
    List<Ad> ads = catalog.getAds(Collections.emptyList());
    assertEquals(AdCatalog.MAX_ADS_TO_SERVE, ads.size());
  }

  @Test
  void fallsBackToRandomAdsForNullContextKeys() {
    AdCatalog catalog = new AdCatalog(new Random(7));
    List<Ad> ads = catalog.getAds(null);
    assertEquals(AdCatalog.MAX_ADS_TO_SERVE, ads.size());
  }

  @Test
  void randomAdsAreNeverEmpty() {
    AdCatalog catalog = new AdCatalog(new Random(3));
    assertFalse(catalog.getRandomAds().isEmpty());
  }
}
