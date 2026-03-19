package com.example.imageprocessing.config;

import java.time.Duration;
import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * Externalized configuration for the image processing API. All values can be overridden via
 * environment variables or application.yml.
 */
@ConfigurationProperties(prefix = "app")
public class AppProperties {

  /** Maximum allowed source image size in bytes. Default: 50 MB. */
  private long maxSourceSize = 50L * 1024 * 1024;

  /** Maximum output image width in pixels. Default: 1400. */
  private int maxOutputWidth = 1400;

  /** Maximum output image height in pixels. Default: 1400. */
  private int maxOutputHeight = 1400;

  /** Cache entry TTL (idle eviction). Default: 5 minutes. */
  private Duration cacheTtl = Duration.ofMinutes(5);

  public long getMaxSourceSize() {
    return maxSourceSize;
  }

  public void setMaxSourceSize(long maxSourceSize) {
    this.maxSourceSize = maxSourceSize;
  }

  public int getMaxOutputWidth() {
    return maxOutputWidth;
  }

  public void setMaxOutputWidth(int maxOutputWidth) {
    this.maxOutputWidth = maxOutputWidth;
  }

  public int getMaxOutputHeight() {
    return maxOutputHeight;
  }

  public void setMaxOutputHeight(int maxOutputHeight) {
    this.maxOutputHeight = maxOutputHeight;
  }

  public Duration getCacheTtl() {
    return cacheTtl;
  }

  public void setCacheTtl(Duration cacheTtl) {
    this.cacheTtl = cacheTtl;
  }
}
