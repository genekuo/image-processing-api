package com.example.imageprocessing.service;

import com.example.imageprocessing.config.AppProperties;
import com.github.benmanes.caffeine.cache.Cache;
import com.github.benmanes.caffeine.cache.Caffeine;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicLong;
import org.springframework.stereotype.Service;

/**
 * Thread-safe in-memory image cache backed by Caffeine.
 *
 * <p>Entries are evicted after they have not been accessed for the configured TTL (idle eviction).
 * Hit and miss counters are maintained for metrics.
 */
@Service
public class ImageCacheService {

  private final Cache<String, byte[]> store;
  private final AtomicLong hits = new AtomicLong();
  private final AtomicLong misses = new AtomicLong();

  public ImageCacheService(AppProperties props) {
    this.store =
        Caffeine.newBuilder().expireAfterAccess(props.getCacheTtl()).build();
  }

  /**
   * Retrieves cached PNG bytes for the given key.
   *
   * @param key cache key
   * @return optional byte array; empty on miss
   */
  public Optional<byte[]> get(String key) {
    byte[] value = store.getIfPresent(key);
    if (value != null) {
      hits.incrementAndGet();
      return Optional.of(value);
    }
    misses.incrementAndGet();
    return Optional.empty();
  }

  /**
   * Stores PNG bytes under the given key.
   *
   * @param key cache key
   * @param data PNG bytes to cache
   */
  public void put(String key, byte[] data) {
    store.put(key, data);
  }

  /** Returns the estimated number of entries currently in the cache. */
  public long size() {
    return store.estimatedSize();
  }

  /** Returns the cumulative number of cache hits. */
  public long hits() {
    return hits.get();
  }

  /** Returns the cumulative number of cache misses. */
  public long misses() {
    return misses.get();
  }

  /** Manually triggers cleanup of expired entries. Useful in tests. */
  public void cleanUp() {
    store.cleanUp();
  }
}
