package com.example.imageprocessing.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.example.imageprocessing.model.Operation;
import java.awt.Color;
import java.awt.image.BufferedImage;
import java.io.IOException;
import java.util.List;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class ImageProcessorServiceTest {

  private ImageProcessorService service;

  @BeforeEach
  void setUp() {
    service = new ImageProcessorService();
  }

  // ── parseOperation ────────────────────────────────────────────────────────

  @Test
  void parseOperation_rotate90() {
    Operation op = service.parseOperation("rotate-90");
    assertThat(op.type()).isEqualTo("rotate");
    assertThat(op.angle()).isEqualTo(90);
  }

  @Test
  void parseOperation_rotate180() {
    Operation op = service.parseOperation("rotate-180");
    assertThat(op.type()).isEqualTo("rotate");
    assertThat(op.angle()).isEqualTo(180);
  }

  @Test
  void parseOperation_rotate270() {
    Operation op = service.parseOperation("rotate-270");
    assertThat(op.type()).isEqualTo("rotate");
    assertThat(op.angle()).isEqualTo(270);
  }

  @Test
  void parseOperation_resize() {
    Operation op = service.parseOperation("resize-800x600");
    assertThat(op.type()).isEqualTo("resize");
    assertThat(op.width()).isEqualTo(800);
    assertThat(op.height()).isEqualTo(600);
  }

  @Test
  void parseOperation_resize_maxDimensions() {
    Operation op = service.parseOperation("resize-1400x1400");
    assertThat(op.width()).isEqualTo(1400);
    assertThat(op.height()).isEqualTo(1400);
  }

  @Test
  void parseOperation_resize_minimal() {
    Operation op = service.parseOperation("resize-1x1");
    assertThat(op.width()).isEqualTo(1);
    assertThat(op.height()).isEqualTo(1);
  }

  @Test
  void parseOperation_trimWhitespace() {
    Operation op = service.parseOperation("  rotate-90  ");
    assertThat(op.type()).isEqualTo("rotate");
    assertThat(op.angle()).isEqualTo(90);
  }

  @Test
  void parseOperation_empty_throws() {
    assertThatThrownBy(() -> service.parseOperation(""))
        .isInstanceOf(IllegalArgumentException.class)
        .hasMessageContaining("empty");
  }

  @Test
  void parseOperation_unknown_throws() {
    assertThatThrownBy(() -> service.parseOperation("flip-horizontal"))
        .isInstanceOf(IllegalArgumentException.class)
        .hasMessageContaining("unknown operation");
  }

  @Test
  void parseOperation_rotate_badAngle_throws() {
    assertThatThrownBy(() -> service.parseOperation("rotate-45"))
        .isInstanceOf(IllegalArgumentException.class);
  }

  @Test
  void parseOperation_resize_missingHeight_throws() {
    assertThatThrownBy(() -> service.parseOperation("resize-800"))
        .isInstanceOf(IllegalArgumentException.class);
  }

  @Test
  void parseOperation_resize_zeroWidth_throws() {
    assertThatThrownBy(() -> service.parseOperation("resize-0x600"))
        .isInstanceOf(IllegalArgumentException.class);
  }

  @Test
  void parseOperation_resize_negativeWidth_throws() {
    assertThatThrownBy(() -> service.parseOperation("resize--1x600"))
        .isInstanceOf(IllegalArgumentException.class);
  }

  @Test
  void parseOperation_resize_nonNumeric_throws() {
    assertThatThrownBy(() -> service.parseOperation("resize-abcxdef"))
        .isInstanceOf(IllegalArgumentException.class);
  }

  @Test
  void parseOperation_resize_exceedsMaxWidth_throws() {
    assertThatThrownBy(() -> service.parseOperation("resize-1401x600"))
        .isInstanceOf(IllegalArgumentException.class)
        .hasMessageContaining("exceed");
  }

  @Test
  void parseOperation_resize_exceedsMaxHeight_throws() {
    assertThatThrownBy(() -> service.parseOperation("resize-600x1401"))
        .isInstanceOf(IllegalArgumentException.class)
        .hasMessageContaining("exceed");
  }

  // ── parseOperations ───────────────────────────────────────────────────────

  @Test
  void parseOperations_commaSeparated() {
    List<Operation> ops = service.parseOperations("rotate-90,resize-200x100,rotate-180");
    assertThat(ops).hasSize(3);
    assertThat(ops.get(0)).isEqualTo(Operation.rotate(90));
    assertThat(ops.get(1)).isEqualTo(Operation.resize(200, 100));
    assertThat(ops.get(2)).isEqualTo(Operation.rotate(180));
  }

  @Test
  void parseOperations_empty_throws() {
    assertThatThrownBy(() -> service.parseOperations(""))
        .isInstanceOf(IllegalArgumentException.class);
  }

  @Test
  void parseOperations_null_throws() {
    assertThatThrownBy(() -> service.parseOperations(null))
        .isInstanceOf(IllegalArgumentException.class);
  }

  // ── rotate ────────────────────────────────────────────────────────────────

  @Test
  void rotate90_swapsDimensions() throws IOException {
    BufferedImage src = makeImage(100, 50);
    BufferedImage result = service.apply(src, Operation.rotate(90));
    assertThat(result.getWidth()).isEqualTo(50);
    assertThat(result.getHeight()).isEqualTo(100);
  }

  @Test
  void rotate180_preservesDimensions() throws IOException {
    BufferedImage src = makeImage(100, 50);
    BufferedImage result = service.apply(src, Operation.rotate(180));
    assertThat(result.getWidth()).isEqualTo(100);
    assertThat(result.getHeight()).isEqualTo(50);
  }

  @Test
  void rotate270_swapsDimensions() throws IOException {
    BufferedImage src = makeImage(100, 50);
    BufferedImage result = service.apply(src, Operation.rotate(270));
    assertThat(result.getWidth()).isEqualTo(50);
    assertThat(result.getHeight()).isEqualTo(100);
  }

  @Test
  void rotate_unsupportedAngle_throws() {
    BufferedImage src = makeImage(10, 10);
    assertThatThrownBy(() -> service.apply(src, Operation.rotate(45)))
        .isInstanceOf(IllegalArgumentException.class)
        .hasMessageContaining("unsupported rotation angle");
  }

  // ── resize ────────────────────────────────────────────────────────────────

  @Test
  void resize_exactDimensions() throws IOException {
    BufferedImage src = makeImage(200, 100);
    BufferedImage result = service.apply(src, Operation.resize(50, 50));
    assertThat(result.getWidth()).isEqualTo(50);
    assertThat(result.getHeight()).isEqualTo(50);
  }

  @Test
  void resize_wideSource_coverCrop() throws IOException {
    BufferedImage src = makeImage(300, 100);
    BufferedImage result = service.apply(src, Operation.resize(100, 100));
    assertThat(result.getWidth()).isEqualTo(100);
    assertThat(result.getHeight()).isEqualTo(100);
  }

  @Test
  void resize_tallSource_coverCrop() throws IOException {
    BufferedImage src = makeImage(100, 400);
    BufferedImage result = service.apply(src, Operation.resize(80, 80));
    assertThat(result.getWidth()).isEqualTo(80);
    assertThat(result.getHeight()).isEqualTo(80);
  }

  // ── apply / applyAll ──────────────────────────────────────────────────────

  @Test
  void apply_unknownType_throws() {
    BufferedImage src = makeImage(10, 10);
    Operation unknown = new Operation("blur", 0, 0, 0);
    assertThatThrownBy(() -> service.apply(src, unknown))
        .isInstanceOf(IllegalArgumentException.class)
        .hasMessageContaining("unsupported operation type");
  }

  @Test
  void applyAll_chainsOperations() throws IOException {
    // 200x100 → rotate-90 → 100x200 → resize-50x50 → 50x50
    BufferedImage src = makeImage(200, 100);
    List<Operation> ops = List.of(Operation.rotate(90), Operation.resize(50, 50));
    BufferedImage result = service.applyAll(src, ops);
    assertThat(result.getWidth()).isEqualTo(50);
    assertThat(result.getHeight()).isEqualTo(50);
  }

  @Test
  void applyAll_emptyOps_returnsOriginal() throws IOException {
    BufferedImage src = makeImage(100, 50);
    BufferedImage result = service.applyAll(src, List.of());
    assertThat(result.getWidth()).isEqualTo(100);
    assertThat(result.getHeight()).isEqualTo(50);
  }

  // ── helpers ───────────────────────────────────────────────────────────────

  private static BufferedImage makeImage(int w, int h) {
    BufferedImage img = new BufferedImage(w, h, BufferedImage.TYPE_INT_RGB);
    var g = img.createGraphics();
    g.setColor(Color.RED);
    g.fillRect(0, 0, w, h);
    g.dispose();
    return img;
  }
}
