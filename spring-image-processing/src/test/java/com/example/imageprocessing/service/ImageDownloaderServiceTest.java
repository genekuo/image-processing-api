package com.example.imageprocessing.service;

import static org.assertj.core.api.Assertions.assertThat;

import com.example.imageprocessing.config.AppProperties;
import java.awt.Color;
import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.util.concurrent.TimeUnit;
import javax.imageio.ImageIO;
import okhttp3.mockwebserver.MockResponse;
import okhttp3.mockwebserver.MockWebServer;
import okio.Buffer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import reactor.test.StepVerifier;

class ImageDownloaderServiceTest {

  private MockWebServer mockServer;
  private ImageDownloaderService service;

  @BeforeEach
  void setUp() throws IOException {
    mockServer = new MockWebServer();
    mockServer.start();

    AppProperties props = new AppProperties();
    props.setMaxSourceSize(50L * 1024 * 1024);
    service = new ImageDownloaderService(props);
  }

  @AfterEach
  void tearDown() throws IOException {
    mockServer.shutdown();
  }

  @Test
  void download_jpeg_success() throws IOException {
    byte[] jpeg = encodeImage(makeTestImage(100, 50), "JPEG");
    mockServer.enqueue(new MockResponse()
        .setResponseCode(200)
        .addHeader("Content-Type", "image/jpeg")
        .setBody(new Buffer().write(jpeg)));

    String url = mockServer.url("/test.jpg").toString();

    StepVerifier.create(service.download(url))
        .assertNext(img -> {
          assertThat(img.getWidth()).isEqualTo(100);
          assertThat(img.getHeight()).isEqualTo(50);
        })
        .verifyComplete();
  }

  @Test
  void download_png_success() throws IOException {
    byte[] png = encodeImage(makeTestImage(80, 60), "PNG");
    mockServer.enqueue(new MockResponse()
        .setResponseCode(200)
        .addHeader("Content-Type", "image/png")
        .setBody(new Buffer().write(png)));

    String url = mockServer.url("/test.png").toString();

    StepVerifier.create(service.download(url))
        .assertNext(img -> {
          assertThat(img.getWidth()).isEqualTo(80);
          assertThat(img.getHeight()).isEqualTo(60);
        })
        .verifyComplete();
  }

  @Test
  void download_gif_success() throws IOException {
    byte[] gif = encodeImage(makeTestImage(50, 50), "GIF");
    mockServer.enqueue(new MockResponse()
        .setResponseCode(200)
        .addHeader("Content-Type", "image/gif")
        .setBody(new Buffer().write(gif)));

    String url = mockServer.url("/test.gif").toString();

    StepVerifier.create(service.download(url))
        .assertNext(img -> assertThat(img).isNotNull())
        .verifyComplete();
  }

  @Test
  void download_maxSizeExceeded_returnsError() throws IOException {
    // A tiny maxSourceSize so the image exceeds it
    AppProperties props = new AppProperties();
    props.setMaxSourceSize(10);
    ImageDownloaderService tinyService = new ImageDownloaderService(props);

    byte[] png = encodeImage(makeTestImage(200, 200), "PNG");
    mockServer.enqueue(new MockResponse()
        .setResponseCode(200)
        .addHeader("Content-Type", "image/png")
        .setBody(new Buffer().write(png)));

    String url = mockServer.url("/big.png").toString();

    StepVerifier.create(tinyService.download(url))
        .expectError()
        .verify();
  }

  @Test
  void download_httpError_returnsError() {
    mockServer.enqueue(new MockResponse().setResponseCode(404).setBody("not found"));

    String url = mockServer.url("/missing.png").toString();

    StepVerifier.create(service.download(url))
        .expectErrorSatisfies(e -> assertThat(e.getMessage()).contains("404"))
        .verify();
  }

  @Test
  void download_serverError_returnsError() {
    mockServer.enqueue(new MockResponse().setResponseCode(500));

    String url = mockServer.url("/error.png").toString();

    StepVerifier.create(service.download(url))
        .expectErrorSatisfies(e -> assertThat(e.getMessage()).contains("500"))
        .verify();
  }

  @Test
  void download_invalidScheme_ftp_returnsError() {
    StepVerifier.create(service.download("ftp://example.com/image.png"))
        .expectErrorSatisfies(e -> {
          assertThat(e).isInstanceOf(IllegalArgumentException.class);
          assertThat(e.getMessage()).contains("Invalid URL scheme");
        })
        .verify();
  }

  @Test
  void download_invalidScheme_noScheme_returnsError() {
    StepVerifier.create(service.download("example.com/image.png"))
        .expectErrorSatisfies(e -> assertThat(e).isInstanceOf(IllegalArgumentException.class))
        .verify();
  }

  @Test
  void download_invalidScheme_dataUri_returnsError() {
    StepVerifier.create(service.download("data:image/png;base64,abc"))
        .expectErrorSatisfies(e -> assertThat(e).isInstanceOf(IllegalArgumentException.class))
        .verify();
  }

  @Test
  void download_timeout_returnsError() throws IOException {
    // Slow server that hangs
    mockServer.enqueue(new MockResponse()
        .setBodyDelay(5, TimeUnit.SECONDS)
        .setResponseCode(200)
        .setBody("never"));

    AppProperties props = new AppProperties();
    props.setMaxSourceSize(50L * 1024 * 1024);
    ImageDownloaderService timeoutService = new ImageDownloaderService(props);

    String url = mockServer.url("/slow.png").toString();

    StepVerifier.withVirtualTime(() -> timeoutService.download(url))
        .thenAwait(java.time.Duration.ofSeconds(31))
        .expectError()
        .verify();
  }

  @Test
  void download_corruptBody_returnsError() {
    mockServer.enqueue(new MockResponse()
        .setResponseCode(200)
        .addHeader("Content-Type", "image/png")
        .setBody("not-a-real-image"));

    String url = mockServer.url("/corrupt.png").toString();

    StepVerifier.create(service.download(url))
        .expectErrorSatisfies(e -> assertThat(e.getMessage()).contains("Failed to decode"))
        .verify();
  }

  // ── helpers ───────────────────────────────────────────────────────────────

  private static BufferedImage makeTestImage(int w, int h) {
    BufferedImage img = new BufferedImage(w, h, BufferedImage.TYPE_INT_RGB);
    var g = img.createGraphics();
    g.setColor(Color.BLUE);
    g.fillRect(0, 0, w, h);
    g.dispose();
    return img;
  }

  private static byte[] encodeImage(BufferedImage img, String format) throws IOException {
    ByteArrayOutputStream baos = new ByteArrayOutputStream();
    ImageIO.write(img, format, baos);
    return baos.toByteArray();
  }
}
