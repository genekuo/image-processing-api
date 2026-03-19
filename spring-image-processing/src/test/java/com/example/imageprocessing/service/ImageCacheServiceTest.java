package com.example.imageprocessing.service;

import static org.assertj.core.api.Assertions.assertThat;

import com.example.imageprocessing.config.AppProperties;
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class ImageCacheServiceTest {

  private ImageCacheService cache;

  @BeforeEach
  void setUp() {
    AppProperties props = new AppProperties();
    props.setCacheTtl(Duration.ofMinutes(5));
    cache = new ImageCacheService(props);
  }

  @Test
  void setAndGet_returnsStoredValue() {
    cache.put("key1", new byte[]{1, 2, 3});
    Optional<byte[]> result = cache.get("key1");
    assertThat(result).isPresent();
    assertThat(result.get()).containsExactly(1, 2, 3);
  }

  @Test
  void get_miss_returnsEmpty() {
    Optional<byte[]> result = cache.get("nonexistent");
    assertThat(result).isEmpty();
  }

  @Test
  void get_hit_incrementsHitCounter() {
    cache.put("k", new byte[]{42});
    cache.get("k");
    cache.get("k");
    assertThat(cache.hits()).isEqualTo(2);
    assertThat(cache.misses()).isEqualTo(0);
  }

  @Test
  void get_miss_incrementsMissCounter() {
    cache.get("missing");
    cache.get("missing2");
    assertThat(cache.misses()).isEqualTo(2);
    assertThat(cache.hits()).isEqualTo(0);
  }

  @Test
  void ttlExpiration_evictsEntry() throws InterruptedException {
    AppProperties props = new AppProperties();
    props.setCacheTtl(Duration.ofMillis(100));
    ImageCacheService shortTtlCache = new ImageCacheService(props);

    shortTtlCache.put("ephemeral", new byte[]{7});
    assertThat(shortTtlCache.get("ephemeral")).isPresent();

    Thread.sleep(200);
    shortTtlCache.cleanUp(); // trigger Caffeine eviction

    assertThat(shortTtlCache.get("ephemeral")).isEmpty();
  }

  @Test
  void get_resetsIdleTtl() throws InterruptedException {
    AppProperties props = new AppProperties();
    props.setCacheTtl(Duration.ofMillis(300));
    ImageCacheService shortTtlCache = new ImageCacheService(props);

    shortTtlCache.put("active", new byte[]{1});

    // Access before TTL expires to reset the idle timer
    Thread.sleep(150);
    assertThat(shortTtlCache.get("active")).isPresent();

    // Another 150ms — total 300ms but TTL was reset, so entry should still be present
    Thread.sleep(150);
    shortTtlCache.cleanUp();
    assertThat(shortTtlCache.get("active")).isPresent();
  }

  @Test
  void size_reflectsEntryCount() {
    assertThat(cache.size()).isEqualTo(0);
    cache.put("a", new byte[]{1});
    cache.put("b", new byte[]{2});
    cache.cleanUp();
    assertThat(cache.size()).isEqualTo(2);
  }

  @Test
  void evictionSelectivity_keepsRecentlyAccessedEntries() throws InterruptedException {
    AppProperties props = new AppProperties();
    props.setCacheTtl(Duration.ofMillis(200));
    ImageCacheService shortTtlCache = new ImageCacheService(props);

    shortTtlCache.put("stale", new byte[]{1});
    shortTtlCache.put("fresh", new byte[]{2});

    Thread.sleep(100);
    // Keep 'fresh' alive by accessing it
    shortTtlCache.get("fresh");

    Thread.sleep(150); // stale is now past 200ms idle, fresh was reset at 100ms
    shortTtlCache.cleanUp();

    assertThat(shortTtlCache.get("stale")).isEmpty();
    assertThat(shortTtlCache.get("fresh")).isPresent();
  }

  @Test
  void concurrentAccess_noRaceConditions() throws InterruptedException {
    int threads = 50;
    int iterations = 100;
    CountDownLatch ready = new CountDownLatch(threads);
    CountDownLatch start = new CountDownLatch(1);
    CountDownLatch done = new CountDownLatch(threads);
    List<Throwable> errors = new ArrayList<>();

    ExecutorService pool = Executors.newFixedThreadPool(threads);
    for (int t = 0; t < threads; t++) {
      final int tid = t;
      pool.submit(() -> {
        ready.countDown();
        try {
          start.await();
          for (int i = 0; i < iterations; i++) {
            String key = "key-" + (tid % 10);
            cache.put(key, new byte[]{(byte) i});
            cache.get(key);
          }
        } catch (Exception e) {
          errors.add(e);
        } finally {
          done.countDown();
        }
      });
    }

    ready.await();
    start.countDown();
    assertThat(done.await(10, TimeUnit.SECONDS)).isTrue();
    pool.shutdown();

    assertThat(errors).isEmpty();
  }
}
