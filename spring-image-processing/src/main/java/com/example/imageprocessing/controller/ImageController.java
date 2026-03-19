package com.example.imageprocessing.controller;

import com.example.imageprocessing.model.Operation;
import com.example.imageprocessing.service.ImageCacheService;
import com.example.imageprocessing.service.ImageDownloaderService;
import com.example.imageprocessing.service.ImageProcessorService;
import com.example.imageprocessing.util.PlaceholderGenerator;
import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.Collections;
import java.util.HexFormat;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import javax.imageio.ImageIO;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import reactor.core.publisher.Mono;
import reactor.core.scheduler.Schedulers;

/**
 * HTTP handlers for the image processing API.
 *
 * <p>Endpoints:
 *
 * <ul>
 *   <li>{@code GET /image?url=...&op=...} — download and transform an image
 *   <li>{@code GET /health} — liveness probe
 *   <li>{@code GET /ready} — readiness probe
 * </ul>
 */
@RestController
public class ImageController {

  private static final Logger log = LoggerFactory.getLogger(ImageController.class);

  private final ImageDownloaderService downloader;
  private final ImageProcessorService processor;
  private final ImageCacheService cache;
  private final PlaceholderGenerator placeholder;

  private final Counter cacheHits;
  private final Counter cacheMisses;
  private final Timer processingTimer;

  public ImageController(
      ImageDownloaderService downloader,
      ImageProcessorService processor,
      ImageCacheService cache,
      PlaceholderGenerator placeholder,
      MeterRegistry registry) {
    this.downloader = downloader;
    this.processor = processor;
    this.cache = cache;
    this.placeholder = placeholder;

    this.cacheHits = Counter.builder("image_cache_hits_total")
        .description("Total image cache hits")
        .register(registry);
    this.cacheMisses = Counter.builder("image_cache_misses_total")
        .description("Total image cache misses")
        .register(registry);
    this.processingTimer = Timer.builder("image_processing_duration_seconds")
        .description("Image processing pipeline duration")
        .register(registry);
  }

  /**
   * Processes an image from a remote URL and returns a PNG.
   *
   * @param url source image URL (http/https)
   * @param opStr comma-separated operations (e.g. {@code rotate-90,resize-400x300})
   * @return PNG image or a color-coded placeholder on error
   */
  @GetMapping("/image")
  public Mono<ResponseEntity<byte[]>> processImage(
      @RequestParam(required = false) String url,
      @RequestParam(name = "op", required = false) String opStr) {

    if (url == null || url.isBlank()) {
      return errorResponse(HttpStatus.BAD_REQUEST, "missing required parameter: url",
          Collections.emptyList());
    }
    if (opStr == null || opStr.isBlank()) {
      return errorResponse(HttpStatus.BAD_REQUEST, "missing required parameter: op",
          Collections.emptyList());
    }

    List<Operation> ops;
    try {
      ops = processor.parseOperations(opStr);
    } catch (IllegalArgumentException e) {
      return errorResponse(HttpStatus.BAD_REQUEST, "invalid operation: " + e.getMessage(),
          Collections.emptyList());
    }

    String key = cacheKey(url, opStr);
    Optional<byte[]> cached = cache.get(key);
    if (cached.isPresent()) {
      cacheHits.increment();
      return Mono.just(pngResponse(HttpStatus.OK, cached.get()));
    }
    cacheMisses.increment();

    final List<Operation> finalOps = ops;
    return downloader
        .download(url)
        .flatMap(
            img ->
                Mono.fromCallable(
                        () -> {
                          Timer.Sample sample = Timer.start();
                          try {
                            var result = processor.applyAll(img, finalOps);
                            ByteArrayOutputStream baos = new ByteArrayOutputStream();
                            ImageIO.write(result, "PNG", baos);
                            return baos.toByteArray();
                          } finally {
                            sample.stop(processingTimer);
                          }
                        })
                    .subscribeOn(Schedulers.boundedElastic()))
        .doOnNext(data -> cache.put(key, data))
        .map(data -> pngResponse(HttpStatus.OK, data))
        .onErrorResume(
            IllegalArgumentException.class,
            e -> {
              log.error("Processing failed: {}", e.getMessage());
              return errorResponse(HttpStatus.INTERNAL_SERVER_ERROR,
                  "processing failed: " + e.getMessage(), finalOps);
            })
        .onErrorResume(
            e -> {
              log.error("Download failed for {}: {}", url, e.getMessage());
              return errorResponse(HttpStatus.BAD_GATEWAY,
                  "download failed: " + e.getMessage(), finalOps);
            });
  }

  /** Liveness probe. Returns {@code {"status":"ok"}}. */
  @GetMapping("/health")
  public Mono<ResponseEntity<Map<String, String>>> health() {
    return Mono.just(ResponseEntity.ok(Map.of("status", "ok")));
  }

  /** Readiness probe. Returns {@code {"status":"ready"}}. */
  @GetMapping("/ready")
  public Mono<ResponseEntity<Map<String, String>>> ready() {
    return Mono.just(ResponseEntity.ok(Map.of("status", "ready")));
  }

  // ── helpers ──────────────────────────────────────────────────────────────

  private Mono<ResponseEntity<byte[]>> errorResponse(
      HttpStatus status, String msg, List<Operation> ops) {
    log.warn("Returning error placeholder: status={} message={}", status, msg);
    int[] dims = extractDimensions(ops);
    try {
      byte[] data = placeholder.generate(status.value(), dims[0], dims[1]);
      return Mono.just(pngResponse(status, data));
    } catch (IOException e) {
      log.error("Failed to generate placeholder: {}", e.getMessage());
      return Mono.just(ResponseEntity.internalServerError().build());
    }
  }

  private ResponseEntity<byte[]> pngResponse(HttpStatus status, byte[] data) {
    return ResponseEntity.status(status)
        .contentType(MediaType.IMAGE_PNG)
        .body(data);
  }

  /**
   * Computes a SHA-256 hex digest of {@code url + "|" + ops} as the cache key.
   */
  String cacheKey(String url, String ops) {
    try {
      MessageDigest digest = MessageDigest.getInstance("SHA-256");
      byte[] hash = digest.digest((url + "|" + ops).getBytes(StandardCharsets.UTF_8));
      return HexFormat.of().formatHex(hash);
    } catch (NoSuchAlgorithmException e) {
      throw new IllegalStateException("SHA-256 not available", e);
    }
  }

  /**
   * Returns the width and height from the last resize operation, or {0, 0} if none.
   */
  int[] extractDimensions(List<Operation> ops) {
    for (int i = ops.size() - 1; i >= 0; i--) {
      if ("resize".equals(ops.get(i).type())) {
        return new int[] {ops.get(i).width(), ops.get(i).height()};
      }
    }
    return new int[] {0, 0};
  }
}
