package com.example.imageprocessing.util;

import java.awt.Color;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import javax.imageio.ImageIO;
import org.springframework.stereotype.Component;

/**
 * Generates color-coded PNG placeholder images for error responses.
 *
 * <ul>
 *   <li>4xx → orange background (#FF8C00)
 *   <li>5xx → red background (#DC143C)
 *   <li>other → gray background (#808080)
 * </ul>
 */
@Component
public class PlaceholderGenerator {

  private static final int DEFAULT_WIDTH = 400;
  private static final int DEFAULT_HEIGHT = 300;
  private static final int MAX_DIMENSION = 1400;

  private static final Color COLOR_ORANGE = new Color(0xFF, 0x8C, 0x00);
  private static final Color COLOR_RED = new Color(0xDC, 0x14, 0x3C);
  private static final Color COLOR_GRAY = new Color(0x80, 0x80, 0x80);

  /**
   * Generates a PNG-encoded placeholder image for the given HTTP status code.
   *
   * @param statusCode HTTP status code to display
   * @param width desired width (0 → default 400)
   * @param height desired height (0 → default 300); clamped to 1400
   * @return PNG bytes
   * @throws IOException if encoding fails
   */
  public byte[] generate(int statusCode, int width, int height) throws IOException {
    int w = clampDimension(width == 0 ? DEFAULT_WIDTH : width);
    int h = clampDimension(height == 0 ? DEFAULT_HEIGHT : height);

    BufferedImage img = new BufferedImage(w, h, BufferedImage.TYPE_INT_RGB);
    Graphics2D g = img.createGraphics();
    try {
      g.setColor(backgroundColor(statusCode));
      g.fillRect(0, 0, w, h);
      drawCenteredText(g, String.valueOf(statusCode), w, h);
    } finally {
      g.dispose();
    }

    ByteArrayOutputStream baos = new ByteArrayOutputStream();
    if (!ImageIO.write(img, "PNG", baos)) {
      throw new IOException("No PNG writer available");
    }
    return baos.toByteArray();
  }

  /**
   * Returns the background color for the given HTTP status code.
   *
   * @param code HTTP status code
   * @return background color
   */
  Color backgroundColor(int code) {
    if (code >= 400 && code < 500) {
      return COLOR_ORANGE;
    }
    if (code >= 500 && code < 600) {
      return COLOR_RED;
    }
    return COLOR_GRAY;
  }

  private void drawCenteredText(Graphics2D g, String text, int width, int height) {
    g.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_ON);
    g.setColor(Color.WHITE);

    // Start at ~40% of image height, scale down if text too wide
    int fontSize = Math.max(8, (int) (height * 0.4));
    Font font = new Font(Font.SANS_SERIF, Font.BOLD, fontSize);
    g.setFont(font);

    FontMetrics fm = g.getFontMetrics();
    int textWidth = fm.stringWidth(text);
    if (textWidth > width * 0.8) {
      fontSize = Math.max(8, (int) (fontSize * (width * 0.8) / textWidth));
      font = new Font(Font.SANS_SERIF, Font.BOLD, fontSize);
      g.setFont(font);
      fm = g.getFontMetrics();
      textWidth = fm.stringWidth(text);
    }

    int x = (width - textWidth) / 2;
    int y = (height + fm.getAscent() - fm.getDescent()) / 2;
    g.drawString(text, x, y);
  }

  private int clampDimension(int value) {
    return Math.min(value, MAX_DIMENSION);
  }
}
