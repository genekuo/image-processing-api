package com.example.imageprocessing;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.EnableConfigurationProperties;

import com.example.imageprocessing.config.AppProperties;

/** Entry point for the Spring Boot WebFlux image processing API. */
@SpringBootApplication
@EnableConfigurationProperties(AppProperties.class)
public class ImageProcessingApplication {

  public static void main(String[] args) {
    SpringApplication.run(ImageProcessingApplication.class, args);
  }
}
