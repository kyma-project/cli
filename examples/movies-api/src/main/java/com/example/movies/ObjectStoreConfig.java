package com.example.movies;

import com.sap.cloud.environment.servicebinding.api.DefaultServiceBindingAccessor;
import com.sap.cloud.environment.servicebinding.api.ServiceBinding;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.s3.S3Client;

import java.net.URI;
import java.util.Map;

@Configuration
public class ObjectStoreConfig {

    @Bean
    public S3Client s3Client() {
        ServiceBinding binding = new DefaultServiceBindingAccessor().getServiceBindings().stream()
                .filter(b -> b.getTags().contains("objectstore"))
                .findFirst()
                .orElseThrow(() -> new IllegalStateException("No Object Store binding found"));

        Map<String, Object> creds = binding.getCredentials();
        return S3Client.builder()
                .region(Region.of((String) creds.get("region")))
                .endpointOverride(URI.create((String) creds.get("endpoint")))
                .credentialsProvider(StaticCredentialsProvider.create(
                        AwsBasicCredentials.create(
                                (String) creds.get("access_key_id"),
                                (String) creds.get("secret_access_key"))))
                .build();
    }

    @Bean
    public String bucketName() {
        ServiceBinding binding = new DefaultServiceBindingAccessor().getServiceBindings().stream()
                .filter(b -> b.getTags().contains("objectstore"))
                .findFirst()
                .orElseThrow();
        return (String) binding.getCredentials().get("bucket");
    }
}
