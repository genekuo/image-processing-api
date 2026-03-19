package com.example.imageprocessing.model;

/**
 * Represents a single image transformation operation.
 *
 * <p>For rotate operations, {@code type} is {@code "rotate"} and {@code angle} is 90, 180, or 270.
 * For resize operations, {@code type} is {@code "resize"} and {@code width}/{@code height} are the
 * target dimensions.
 */
public record Operation(String type, int angle, int width, int height) {

  /** Creates a rotate operation. */
  public static Operation rotate(int angle) {
    return new Operation("rotate", angle, 0, 0);
  }

  /** Creates a resize operation. */
  public static Operation resize(int width, int height) {
    return new Operation("resize", 0, width, height);
  }
}
