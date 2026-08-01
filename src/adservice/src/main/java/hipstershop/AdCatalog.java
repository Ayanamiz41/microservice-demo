package hipstershop;

import hipstershop.AdOuterClass.Ad;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Random;
import org.springframework.stereotype.Component;

/**
 * In-memory ad catalog for the Online Boutique AdService.
 *
 * <p>Mirrors the upstream GoogleCloudPlatform/microservices-demo adservice: ads are keyed by
 * product category; when no (or no matching) context keys are provided, a fixed number of random
 * ads is served.
 */
@Component
public final class AdCatalog {

  /** Maximum number of ads served for random / fallback requests (upstream value). */
  static final int MAX_ADS_TO_SERVE = 2;

  private final Map<String, List<Ad>> adsByCategory = new LinkedHashMap<>();
  private final List<Ad> allAds = new ArrayList<>();
  private final Random random;

  public AdCatalog() {
    this(new Random());
  }

  /** Visible for testing: inject a seeded {@link Random} for deterministic results. */
  AdCatalog(Random random) {
    this.random = random;
    put("clothing", ad("/product/66VCHSJNUP", "Tank top for sale. 20% off."));
    put("accessories", ad("/product/1YMWWN1N4O", "Watch for sale. Buy one, get second kit for free"));
    put("footwear", ad("/product/L9ECAV7KIM", "Loafers for sale. Buy one, get second one for free"));
    put("hair", ad("/product/2ZYFJ3GM2N", "Hairdryer for sale. 50% off."));
    put("decor", ad("/product/0PUK6V6EV0", "Candle holder for sale. 30% off."));
    put(
        "kitchen",
        ad("/product/9SIQT8TOJO", "Bamboo glass jar for sale. 10% off."),
        ad("/product/6E92ZMYYFZ", "Mug for sale. Buy two, get third one for free"));
  }

  /**
   * Returns ads for the given context keys, mirroring upstream {@code AdServiceImpl.getAds}:
   * matching category ads are collected; an empty result falls back to random ads; an empty key
   * list returns random ads.
   */
  public List<Ad> getAds(List<String> contextKeys) {
    if (contextKeys == null || contextKeys.isEmpty()) {
      return getRandomAds();
    }
    List<Ad> ads = new ArrayList<>();
    for (String key : contextKeys) {
      List<Ad> categoryAds = adsByCategory.get(key);
      if (categoryAds != null) {
        ads.addAll(categoryAds);
      }
    }
    return ads.isEmpty() ? getRandomAds() : ads;
  }

  /** Returns up to {@link #MAX_ADS_TO_SERVE} random ads (drawn with replacement, as upstream). */
  public List<Ad> getRandomAds() {
    List<Ad> ads = new ArrayList<>(MAX_ADS_TO_SERVE);
    for (int i = 0; i < MAX_ADS_TO_SERVE; i++) {
      ads.add(allAds.get(random.nextInt(allAds.size())));
    }
    return ads;
  }

  private void put(String category, Ad... ads) {
    List<Ad> list = new ArrayList<>(ads.length);
    for (Ad ad : ads) {
      list.add(ad);
      allAds.add(ad);
    }
    adsByCategory.put(category, list);
  }

  private static Ad ad(String redirectUrl, String text) {
    return Ad.newBuilder().setRedirectUrl(redirectUrl).setText(text).build();
  }
}
