package com.example.movies;

import com.fasterxml.jackson.databind.ObjectMapper;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.server.ResponseStatusException;
import software.amazon.awssdk.core.sync.RequestBody;
import software.amazon.awssdk.services.s3.S3Client;
import software.amazon.awssdk.services.s3.model.*;

import java.io.IOException;
import java.util.List;

@RestController
@RequestMapping("/movies")
@Tag(name = "Movies", description = "CRUD operations for movie resources")
public class MovieController {

    private final S3Client s3;
    private final String bucket;
    private final ObjectMapper mapper = new ObjectMapper();

    public MovieController(S3Client s3, String bucketName) {
        this.s3 = s3;
        this.bucket = bucketName;
    }

    @GetMapping
    @Operation(summary = "List all movies")
    public List<Movie> list() throws IOException {
        ListObjectsV2Response response = s3.listObjectsV2(b -> b.bucket(bucket).prefix("movies/"));
        return response.contents().stream()
                .map(obj -> getMovie(obj.key()))
                .toList();
    }

    @GetMapping("/{id}")
    @Operation(summary = "Get a movie by ID")
    public Movie get(@PathVariable String id) {
        return getMovie("movies/" + id + ".json");
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    @Operation(summary = "Create a new movie")
    public Movie create(@RequestBody Movie movie) throws Exception {
        Movie saved = movie.withId(String.valueOf(System.currentTimeMillis()));
        putMovie(saved);
        return saved;
    }

    @PutMapping("/{id}")
    @Operation(summary = "Update an existing movie")
    public Movie update(@PathVariable String id, @RequestBody Movie movie) throws Exception {
        Movie saved = movie.withId(id);
        putMovie(saved);
        return saved;
    }

    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    @Operation(summary = "Delete a movie")
    public void delete(@PathVariable String id) {
        s3.deleteObject(b -> b.bucket(bucket).key("movies/" + id + ".json"));
    }

    private void putMovie(Movie movie) throws Exception {
        byte[] json = mapper.writeValueAsBytes(movie);
        s3.putObject(b -> b.bucket(bucket).key("movies/" + movie.id() + ".json")
                .contentType("application/json"), RequestBody.fromBytes(json));
    }

    private Movie getMovie(String key) {
        try {
            byte[] data = s3.getObject(b -> b.bucket(bucket).key(key)).readAllBytes();
            return mapper.readValue(data, Movie.class);
        } catch (NoSuchKeyException e) {
            throw new ResponseStatusException(HttpStatus.NOT_FOUND, "Movie not found");
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }
}
