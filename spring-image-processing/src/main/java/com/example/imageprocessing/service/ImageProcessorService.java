package com.example.imageprocessing.service;

import com.example.imageprocessing.model.Operation;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import java.awt.geom.AffineTransform;
import java.awt.image.BufferedImage;
import java.io.IOException;
import java.util.Arrays;
import java.util.List;
import net.coobird.thumbnailator.Thumbnails;
import net.coobird.thumbnailator.geometry.Positions;
import org.springframework.stereotype.Service;

/**
 * Parses and applies image transformation operations (rotate, resize).
 *
 * <p>Supported operations:
 *
 * <ul>
 *   <li>{@code rotate-90}, {@code rotate-180}, {@code rotate-270}
 *   <li>{@code resize-WxH} (e.g. {@code resize-800x600}), max 1400×1400
 * </ul>
 */
@Service
public class ImageProcessorService {

  static final int MAX_OUTPUT_WIDTH = 1400;
  static final int MAX_OUTPUT_HEIGHT = 1400;

  /**
   * Parses a single operation string.
   *
   * @param op operation string such as {@code "rotate-90"} or {@code "resize-800x600"}
   * @return parsed Operation
   * @throws IllegalArgumentException on invalid input
   */
  public Operation parseOperation(String op) {
    op = op.trim();
    if (op.isEmpty()) {
      throw new IllegalArgumentException("empty operation");
    }
    return switch (op) {
      case "rotate-90" -> Operation.rotate(90);
      case "rotate-180" -> Operation.rotate(180);
      case "rotate-270" -> Operation.rotate(270);
      default -> {
        if (op.startsWith("resize-")) {
          yield parseResize(op);
        }
        throw new IllegalArgumentException("unknown operation: \"" + op + "\"");
      }
    };
  }

  private Operation parseResize(String op) {
    String dims = op.substring("resize-".length());
    String[] parts = dims.split("x", 2);
    if (parts.length != 2) {
      throw new IllegalArgumentException(
          "invalid resize format \"" + op + "\": expected resize-WxH");
    }
    int w;
    int h;
    try {
      w = Integer.parseInt(parts[0]);
      h = Integer.parseInt(parts[1]);
    } catch (NumberFormatException e) {
      throw new IllegalArgumentException(
          "invalid resize dimensions in \"" + op + "\": must be positive integers");
    }
    if (w <= 0 || h <= 0) {
      throw new IllegalArgumentException(
          "invalid resize dimensions in \"" + op + "\": must be positive integers");
    }
    if (w > MAX_OUTPUT_WIDTH || h > MAX_OUTPUT_HEIGHT) {
      throw new IllegalArgumentException(
          "resize dimensions "
              + w
              + "x"
              + h
              + " exceed maximum allowed "
              + MAX_OUTPUT_WIDTH
              + "x"
              + MAX_OUTPUT_HEIGHT);
    }
    return Operation.resize(w, h);
  }

  /**
   * Parses a comma-separated list of operation strings.
   *
   * @param ops comma-separated operations, e.g. {@code "rotate-90,resize-200x100"}
   * @return list of parsed Operations
   * @throws IllegalArgumentException on invalid input
   */
  public List<Operation> parseOperations(String ops) {
    if (ops == null || ops.trim().isEmpty()) {
      throw new IllegalArgumentException("empty operations string");
    }
    return Arrays.stream(ops.split(",")).map(this::parseOperation).toList();
  }

  /**
   * Applies a single operation to the image.
   *
   * @param img source image
   * @param op operation to apply
   * @return transformed image
   * @throws IllegalArgumentException for unsupported operation types
   * @throws IOException if image processing fails
   */
  public BufferedImage apply(BufferedImage img, Operation op) throws IOException {
    return switch (op.type()) {
      case "rotate" -> applyRotate(img, op.angle());
      case "resize" -> applyResize(img, op.width(), op.height());
      default ->
          throw new IllegalArgumentException("unsupported operation type: \"" + op.type() + "\"");
    };
  }

  private BufferedImage applyRotate(BufferedImage src, int angle) {
    if (angle != 90 && angle != 180 && angle != 270) {
      throw new IllegalArgumentException("unsupported rotation angle: " + angle);
    }
    int srcW = src.getWidth();
    int srcH = src.getHeight();
    int dstW = (angle == 180) ? srcW : srcH;
    int dstH = (angle == 180) ? srcH : srcW;

    BufferedImage result = new BufferedImage(dstW, dstH, BufferedImage.TYPE_INT_ARGB);
    Graphics2D g = result.createGraphics();
    g.setRenderingHint(RenderingHints.KEY_INTERPOLATION,
        RenderingHints.VALUE_INTERPOLATION_BILINEAR);

    AffineTransform at = new AffineTransform();
    switch (angle) {
      case 90 -> {
        at.translate(dstW, 0);
        at.rotate(Math.toRadians(90));
      }
      case 180 -> {
        at.translate(dstW, dstH);
        at.rotate(Math.toRadians(180));
      }
      case 270 -> {
        at.translate(0, dstH);
        at.rotate(Math.toRadians(270));
      }
      default -> throw new IllegalArgumentException("unsupported rotation angle: " + angle);
    }

    g.setTransform(at);
    g.drawImage(src, 0, 0, null);
    g.dispose();
    return result;
  }

  private BufferedImage applyResize(BufferedImage src, int width, int height) throws IOException {
    return Thumbnails.of(src)
        .size(width, height)
        .crop(Positions.CENTER)
        .imageType(BufferedImage.TYPE_INT_ARGB)
        .asBufferedImage();
  }

  /**
   * Applies all operations in sequence.
   *
   * @param img source image
   * @param ops ordered list of operations
   * @return final transformed image
   * @throws IOException if any operation fails
   */
  public BufferedImage applyAll(BufferedImage img, List<Operation> ops) throws IOException {
    for (int i = 0; i < ops.size(); i++) {
      Operation op = ops.get(i);
      try {
        img = apply(img, op);
      } catch (IllegalArgumentException e) {
        throw new IllegalArgumentException("operation " + i + " (" + op.type() + "): " + e.getMessage(), e);
      }
    }
    return img;
  }
}
