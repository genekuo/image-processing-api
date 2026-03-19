package com.example.imageprocessing.service;

import com.example.imageprocessing.config.AppProperties;
import java.awt.image.BufferedImage;
import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.time.Duration;
import javax.imageio.ImageIO;
import org.apache.commons.io.input.BoundedInputStream;
import org.springframework.http.HttpStatusCode;
import org.springframework.stereotype.Service;
import org.springframework.web.reactive.function.client.WebClient;
import reactor.core.publisher.Mono;
import reactor.core.scheduler.Schedulers;

/**
 * Downloads remote images via HTTP/HTTPS and decodes them into {@link BufferedImage}.
 *
 * <p>Only {@code http://} and {@code https://} URLs are accepted. Responses larger than {@code
 * maxSourceSize} bytes are rejected. Decoding runs on a bounded-elastic thread pool since
 * {@link ImageIO} is a blocking operation.
 */
@Service
public class ImageDownloaderService {

  private final WebClient webClient;
  private final long maxSourceSize;

  public ImageDownloaderService(AppProperties props) {
    this.maxSourceSize = props.getMaxSourceSize();
    this.webClient =
        WebClient.builder()
            .codecs(
                cfg ->
                    cfg.defaultCodecs()
                        .maxInMemorySize((int) Math.min(maxSourceSize + 1, Integer.MAX_VALUE)))
            .build();
  }

  /**
   * Downloads the image at {@code rawUrl} and decodes it.
   *
   * @param rawUrl URL of the source image (http/https only)
   * @return Mono of the decoded image
   */
  public Mono<BufferedImage> download(String rawUrl) {
    try {
      validateUrl(rawUrl);
    } catch (IllegalArgumentException e) {
      return Mono.error(e);
    }

    return webClient
        .get()
        .uri(rawUrl)
        .retrieve()
        .onStatus(
            HttpStatusCode::isError,
            resp ->
                Mono.error(
                    new IOException("Unexpected HTTP status " + resp.statusCode() + " for " + rawUrl)))
        .bodyToMono(byte[].class)
        .timeout(Duration.ofSeconds(30))
        .flatMap(bytes -> decodeImage(bytes, rawUrl));
  }

  private Mono<BufferedImage> decodeImage(byte[] bytes, String rawUrl) {
    return Mono.fromCallable(
            () -> {
              if (bytes.length > maxSourceSize) {
                throw new IOException(
                    "Response body exceeds maximum allowed size of " + maxSourceSize + " bytes");
              }
              // Use BoundedInputStream as an extra safety net during decoding
              try (BoundedInputStream bounded =
                  BoundedInputStream.builder()
                      .setInputStream(new ByteArrayInputStream(bytes))
                      .setMaxCount(maxSourceSize + 1)
                      .get()) {
                BufferedImage img = ImageIO.read(bounded);
                if (img == null) {
                  throw new IOException(
                      "Failed to decode image from " + rawUrl + ": unsupported or corrupt format");
                }
                return img;
              }
            })
        .subscribeOn(Schedulers.boundedElastic());
  }

  private void validateUrl(String rawUrl) {
    if (rawUrl == null) {
      throw new IllegalArgumentException("URL must not be null");
    }
    String lower = rawUrl.toLowerCase();
    if (!lower.startsWith("http://") && !lower.startsWith("https://")) {
      throw new IllegalArgumentException(
          "Invalid URL scheme: only http and https are supported, got: \"" + rawUrl + "\"");
    }
  }
}
