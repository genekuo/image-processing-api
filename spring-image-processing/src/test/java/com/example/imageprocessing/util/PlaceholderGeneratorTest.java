package com.example.imageprocessing.util;

import static org.assertj.core.api.Assertions.assertThat;

import java.awt.image.BufferedImage;
import java.io.ByteArrayInputStream;
import java.io.IOException;
import javax.imageio.ImageIO;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;

class PlaceholderGeneratorTest {

  private PlaceholderGenerator generator;

  @BeforeEach
  void setUp() {
    generator = new PlaceholderGenerator();
  }

  @ParameterizedTest
  @ValueSource(ints = {400, 401, 403, 404, 422, 429, 499})
  void generate_4xx_orangeBackground(int code) throws IOException {
    BufferedImage img = generateAndDecode(code, 100, 100);
    int[] rgb = getRgbAt(img, 0, 0);
    assertThat(rgb[0]).isEqualTo(0xFF); // R
    assertThat(rgb[1]).isEqualTo(0x8C); // G
    assertThat(rgb[2]).isEqualTo(0x00); // B
  }

  @ParameterizedTest
  @ValueSource(ints = {500, 502, 503, 504})
  void generate_5xx_redBackground(int code) throws IOException {
    BufferedImage img = generateAndDecode(code, 100, 100);
    int[] rgb = getRgbAt(img, 0, 0);
    assertThat(rgb[0]).isEqualTo(0xDC); // R
    assertThat(rgb[1]).isEqualTo(0x14); // G
    assertThat(rgb[2]).isEqualTo(0x3C); // B
  }

  @ParameterizedTest
  @ValueSource(ints = {200, 301, 302, 600, 0})
  void generate_otherCodes_grayBackground(int code) throws IOException {
    BufferedImage img = generateAndDecode(code, 100, 100);
    int[] rgb = getRgbAt(img, 0, 0);
    assertThat(rgb[0]).isEqualTo(0x80); // R
    assertThat(rgb[1]).isEqualTo(0x80); // G
    assertThat(rgb[2]).isEqualTo(0x80); // B
  }

  @Test
  void generate_zeroDimensions_usesDefaults() throws IOException {
    BufferedImage img = generateAndDecode(404, 0, 0);
    assertThat(img.getWidth()).isEqualTo(400);
    assertThat(img.getHeight()).isEqualTo(300);
  }

  @Test
  void generate_negativeDimensions_usesDefaults() throws IOException {
    BufferedImage img = generateAndDecode(500, -10, -10);
    assertThat(img.getWidth()).isEqualTo(400);
    assertThat(img.getHeight()).isEqualTo(300);
  }

  @Test
  void generate_oversizedDimensions_clamped() throws IOException {
    BufferedImage img = generateAndDecode(500, 5000, 3000);
    assertThat(img.getWidth()).isEqualTo(1400);
    assertThat(img.getHeight()).isEqualTo(1400);
  }

  @Test
  void generate_outputIsValidPng() throws IOException {
    byte[] data = generator.generate(404, 200, 150);
    assertThat(data).isNotEmpty();
    // PNG magic bytes
    assertThat(data[0] & 0xFF).isEqualTo(0x89);
    assertThat(data[1] & 0xFF).isEqualTo(0x50); // 'P'
    assertThat(data[2] & 0xFF).isEqualTo(0x4E); // 'N'
    assertThat(data[3] & 0xFF).isEqualTo(0x47); // 'G'
  }

  @Test
  void generate_tinyImage_doesNotThrow() throws IOException {
    // Tests the min-font-size branch in drawCenteredText
    byte[] data = generator.generate(404, 1, 1);
    assertThat(data).isNotEmpty();
  }

  @Test
  void backgroundColor_4xx_returnsOrange() {
    var color = generator.backgroundColor(404);
    assertThat(color.getRed()).isEqualTo(0xFF);
    assertThat(color.getGreen()).isEqualTo(0x8C);
    assertThat(color.getBlue()).isEqualTo(0x00);
  }

  @Test
  void backgroundColor_5xx_returnsRed() {
    var color = generator.backgroundColor(500);
    assertThat(color.getRed()).isEqualTo(0xDC);
    assertThat(color.getGreen()).isEqualTo(0x14);
    assertThat(color.getBlue()).isEqualTo(0x3C);
  }

  @Test
  void backgroundColor_other_returnsGray() {
    var color = generator.backgroundColor(200);
    assertThat(color.getRed()).isEqualTo(0x80);
    assertThat(color.getGreen()).isEqualTo(0x80);
    assertThat(color.getBlue()).isEqualTo(0x80);
  }

  // ── helpers ───────────────────────────────────────────────────────────────

  private BufferedImage generateAndDecode(int code, int w, int h) throws IOException {
    byte[] data = generator.generate(code, w, h);
    return ImageIO.read(new ByteArrayInputStream(data));
  }

  private int[] getRgbAt(BufferedImage img, int x, int y) {
    int argb = img.getRGB(x, y);
    return new int[]{
        (argb >> 16) & 0xFF,  // R
        (argb >> 8) & 0xFF,   // G
        argb & 0xFF           // B
    };
  }
}
