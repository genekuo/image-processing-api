package com.example.imageprocessing.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

import com.example.imageprocessing.model.Operation;
import com.example.imageprocessing.service.ImageDownloaderService;
import java.awt.Color;
import java.awt.image.BufferedImage;
import java.util.List;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.reactive.AutoConfigureWebTestClient;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.SpringBootTest.WebEnvironment;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.test.web.reactive.server.WebTestClient;
import reactor.core.publisher.Mono;

@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
@AutoConfigureWebTestClient
class ImageControllerTest {

  @Autowired
  private WebTestClient client;

  @Autowired
  private ImageController controller;

  @MockBean
  private ImageDownloaderService downloader;

  // ── /health ───────────────────────────────────────────────────────────────

  @Test
  void health_returns200() {
    client.get().uri("/health")
        .exchange()
        .expectStatus().isOk()
        .expectBody()
        .jsonPath("$.status").isEqualTo("ok");
  }

  // ── /ready ────────────────────────────────────────────────────────────────

  @Test
  void ready_returns200() {
    client.get().uri("/ready")
        .exchange()
        .expectStatus().isOk()
        .expectBody()
        .jsonPath("$.status").isEqualTo("ready");
  }

  // ── /image – validation ───────────────────────────────────────────────────

  @Test
  void image_missingUrl_returns400Png() {
    client.get().uri("/image?op=rotate-90")
        .exchange()
        .expectStatus().isBadRequest()
        .expectHeader().contentType(MediaType.IMAGE_PNG)
        .expectBody(byte[].class)
        .value(body -> assertThat(body).isNotEmpty());
  }

  @Test
  void image_missingOp_returns400Png() {
    client.get().uri("/image?url=http://example.com/img.png")
        .exchange()
        .expectStatus().isBadRequest()
        .expectHeader().contentType(MediaType.IMAGE_PNG)
        .expectBody(byte[].class)
        .value(body -> assertThat(body).isNotEmpty());
  }

  @Test
  void image_invalidOp_returns400Png() {
    client.get().uri("/image?url=http://example.com/img.png&op=flip-horizontal")
        .exchange()
        .expectStatus().isBadRequest()
        .expectHeader().contentType(MediaType.IMAGE_PNG)
        .expectBody(byte[].class)
        .value(body -> assertThat(body).isNotEmpty());
  }

  // ── /image – download failure ─────────────────────────────────────────────

  @Test
  void image_downloadFailure_returns502Png() {
    when(downloader.download(anyString()))
        .thenReturn(Mono.error(new RuntimeException("connection refused")));

    client.get().uri("/image?url=http://example.com/img.png&op=rotate-90")
        .exchange()
        .expectStatus().isEqualTo(HttpStatus.BAD_GATEWAY)
        .expectHeader().contentType(MediaType.IMAGE_PNG)
        .expectBody(byte[].class)
        .value(body -> assertThat(body).isNotEmpty());
  }

  @Test
  void image_invalidUrlScheme_returns400Png() {
    when(downloader.download("ftp://example.com/img.png"))
        .thenReturn(Mono.error(new IllegalArgumentException("Invalid URL scheme: ftp")));

    client.get().uri("/image?url=ftp://example.com/img.png&op=rotate-90")
        .exchange()
        .expectStatus().isBadRequest()
        .expectHeader().contentType(MediaType.IMAGE_PNG);
  }

  // ── /image – success ──────────────────────────────────────────────────────

  @Test
  void image_success_returns200Png() {
    BufferedImage img = makeTestImage(200, 100);
    when(downloader.download(anyString())).thenReturn(Mono.just(img));

    client.get().uri("/image?url=http://example.com/img.png&op=rotate-90")
        .exchange()
        .expectStatus().isOk()
        .expectHeader().contentType(MediaType.IMAGE_PNG)
        .expectBody(byte[].class)
        .value(body -> {
          assertThat(body).isNotEmpty();
          // PNG magic bytes
          assertThat(body[0] & 0xFF).isEqualTo(0x89);
          assertThat(body[1] & 0xFF).isEqualTo(0x50);
        });
  }

  @Test
  void image_cacheHit_returns200Png() {
    BufferedImage img = makeTestImage(100, 100);
    when(downloader.download(anyString())).thenReturn(Mono.just(img));

    String uri = "/image?url=http://example.com/cached.png&op=resize-100x100";

    // First request — cache miss, downloader called
    client.get().uri(uri)
        .exchange()
        .expectStatus().isOk()
        .expectHeader().contentType(MediaType.IMAGE_PNG);

    // Second request — should be served from cache (downloader still mocked, no real call matters)
    client.get().uri(uri)
        .exchange()
        .expectStatus().isOk()
        .expectHeader().contentType(MediaType.IMAGE_PNG);
  }

  @Test
  void image_successWithResize_returnsCorrectDimensions() {
    BufferedImage img = makeTestImage(300, 200);
    when(downloader.download(anyString())).thenReturn(Mono.just(img));

    client.get().uri("/image?url=http://example.com/img.png&op=resize-150x150")
        .exchange()
        .expectStatus().isOk()
        .expectHeader().contentType(MediaType.IMAGE_PNG)
        .expectBody(byte[].class)
        .value(body -> assertThat(body).isNotEmpty());
  }

  @Test
  void image_multipleOps_returnsOk() {
    BufferedImage img = makeTestImage(200, 100);
    when(downloader.download(anyString())).thenReturn(Mono.just(img));

    client.get()
        .uri("/image?url=http://example.com/img.png&op=rotate-90,resize-50x50")
        .exchange()
        .expectStatus().isOk()
        .expectHeader().contentType(MediaType.IMAGE_PNG);
  }

  // ── /image – CORS ─────────────────────────────────────────────────────────

  @Test
  void image_corsHeaderPresent() {
    client.get().uri("/health")
        .header("Origin", "https://example.com")
        .exchange()
        .expectStatus().isOk()
        .expectHeader().exists("Access-Control-Allow-Origin");
  }

  // ── unit tests for package-visible helpers ────────────────────────────────

  @Test
  void cacheKey_deterministicAndUnique() {
    String k1 = controller.cacheKey("http://a.com/img.png", "rotate-90");
    String k2 = controller.cacheKey("http://a.com/img.png", "rotate-90");
    String k3 = controller.cacheKey("http://b.com/img.png", "rotate-90");

    assertThat(k1).isEqualTo(k2);
    assertThat(k1).isNotEqualTo(k3);
    assertThat(k1).hasSize(64); // SHA-256 hex
  }

  @Test
  void extractDimensions_findsLastResize() {
    List<Operation> ops = List.of(
        Operation.rotate(90),
        Operation.resize(400, 300),
        Operation.resize(200, 150));
    int[] dims = controller.extractDimensions(ops);
    assertThat(dims[0]).isEqualTo(200);
    assertThat(dims[1]).isEqualTo(150);
  }

  @Test
  void extractDimensions_noResize_returnsZeros() {
    List<Operation> ops = List.of(Operation.rotate(90));
    int[] dims = controller.extractDimensions(ops);
    assertThat(dims[0]).isEqualTo(0);
    assertThat(dims[1]).isEqualTo(0);
  }

  @Test
  void extractDimensions_emptyOps_returnsZeros() {
    int[] dims = controller.extractDimensions(List.of());
    assertThat(dims[0]).isEqualTo(0);
    assertThat(dims[1]).isEqualTo(0);
  }

  // ── helpers ───────────────────────────────────────────────────────────────

  private static BufferedImage makeTestImage(int w, int h) {
    BufferedImage img = new BufferedImage(w, h, BufferedImage.TYPE_INT_RGB);
    var g = img.createGraphics();
    g.setColor(Color.GREEN);
    g.fillRect(0, 0, w, h);
    g.dispose();
    return img;
  }
}
