package com.example.imageprocessing.config;

import org.springframework.context.annotation.Configuration;
import org.springframework.web.reactive.config.CorsRegistry;
import org.springframework.web.reactive.config.WebFluxConfigurer;

/** Global WebFlux configuration: CORS, codec limits, etc. */
@Configuration
public class WebConfig implements WebFluxConfigurer {

  @Override
  public void addCorsMappings(CorsRegistry registry) {
    registry
        .addMapping("/**")
        .allowedOrigins("*")
        .allowedMethods("GET", "OPTIONS")
        .allowedHeaders("*");
  }
}
