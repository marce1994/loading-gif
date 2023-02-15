// prompt

// necesito una api en go con los siguientes requisitos:

// 1: un endpoint para obtener la url de un gif animado en base a likes y dislikes
// 2: un endpoint para dar like a un gif
// 3: un endpoint para dar dislike a un gif
// 4: los datos se guardaran en redis, y la url del mismo es "redis:6379"
// 5: el endpoint para obtener la url de los gif retorna un gif random, pero los que tienen mas likes tienen mas chance de salir elegidos.
// 6: debera inicializarse data de pruebas para poder testear la api.
// 7: el get de los gifs tambien debe aceptar un parametro opcional que sea el tag del gif para poder filtrar.
// 8: los gifs pueden tener varios tags como por ejemplo Pokemon, Bailando, Volando

// solo necesito el codigo, no expliques nada.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

type Gif struct {
	ID       int      `json:"id"`
	URL      string   `json:"url"`
	Likes    int      `json:"likes"`
	Dislikes int      `json:"dislikes"`
	Tags     []string `json:"tags"`
}

var gifs []Gif

var redisClient *redis.Client

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	router := mux.NewRouter()

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	// Inicializar datos de prueba
	initData()

	// Endpoints
	router.HandleFunc("/", serveIndex).Methods("GET")
	router.HandleFunc("/gif", getRandomGif).Methods("GET")
	router.HandleFunc("/gif/{id}/like", likeGif).Methods("POST")
	router.HandleFunc("/gif/{id}/dislike", dislikeGif).Methods("POST")

	logMiddleware := NewLogMiddleware(logger)
    router.Use(logMiddleware.Func())

	log.Fatal(http.ListenAndServe(":8080", router))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// func getRandomGif(w http.ResponseWriter, r *http.Request) {
// 	var filteredGifs []Gif
// 	tag := r.URL.Query().Get("tag")
// 	if tag != "" {
// 		for _, gif := range gifs {
// 			if contains(gif.Tags, tag) {
// 				filteredGifs = append(filteredGifs, gif)
// 			}
// 		}
// 	} else {
// 		filteredGifs = gifs
// 	}

// 	totalLikes := 0
// 	for _, gif := range filteredGifs {
// 		totalLikes += gif.Likes
// 	}

// 	if totalLikes == 0 {
// 		json.NewEncoder(w).Encode(filteredGifs[rand.Intn(len(filteredGifs))].URL)
// 	} else {
// 		randomNum := rand.Intn(totalLikes)
// 		for _, gif := range filteredGifs {
// 			if randomNum < gif.Likes {
// 				json.NewEncoder(w).Encode(gif.URL)
// 				return
// 			}
// 			randomNum -= gif.Likes
// 		}
// 	}
// }

func getRandomGif(w http.ResponseWriter, r *http.Request) {
    var filteredGifs []Gif
    tag := r.URL.Query().Get("tag")
    if tag != "" {
        for _, gif := range gifs {
            if contains(gif.Tags, tag) {
                filteredGifs = append(filteredGifs, gif)
            }
        }
    } else {
        filteredGifs = gifs
    }
    totalLikes := 0
    for _, gif := range filteredGifs {
        totalLikes += gif.Likes
    }
    if totalLikes == 0 {
        // Descargar imagen usando la URL obtenida
        resp, err := http.Get(filteredGifs[rand.Intn(len(filteredGifs))].URL)
        if err != nil {
            log.Fatalln(err)
        }
        defer resp.Body.Close()
        // Copiar la imagen al response
        _, err = io.Copy(w, resp.Body)
        if err != nil {
            log.Fatalln(err)
        }
    } else {
        randomNum := rand.Intn(totalLikes)
        for _, gif := range filteredGifs {
            if randomNum < gif.Likes {
                // Descargar imagen usando la URL obtenida
                resp, err := http.Get(gif.URL)
                if err != nil {
                    log.Fatalln(err)
                }
                defer resp.Body.Close()
                // Copiar la imagen al response
                _, err = io.Copy(w, resp.Body)
                if err != nil {
                    log.Fatalln(err)
                }
                return
            }
            randomNum -= gif.Likes
        }
    }
}

func likeGif(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	gif, err := getGifByID(id)
	if err != nil {
		http.Error(w, "Gif not found", http.StatusNotFound)
		return
	}

	gif.Likes++
	updateGif(gif)
}

func dislikeGif(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	gif, err := getGifByID(id)
	if err != nil {
		http.Error(w, "Gif not found", http.StatusNotFound)
		return
	}

	gif.Dislikes++
	updateGif(gif)
}

func getGifByID(id string) (*Gif, error) {
	for i := range gifs {
		if strconv.Itoa(gifs[i].ID) == id {
			return &gifs[i], nil
		}
	}

	return nil, errGifNotFound
}

func updateGif(gif *Gif) {
	serialized, _ := json.Marshal(gif)
	redisClient.Set("gif:"+strconv.Itoa(gif.ID), serialized, 0)
}

func initData() {
	gifs = []Gif{
		Gif{ID: 1, URL: "https://media.tenor.com/rt2qSDNvVEQAAAAi/pikachu-dance.gif", Likes: 10, Dislikes: 2, Tags: []string{"Pokemon", "Bailando"}},
		Gif{ID: 2, URL: "https://media.tenor.com/SWzRKMsofdcAAAAi/eevee-dance.gif", Likes: 5, Dislikes: 1, Tags: []string{"Pokemon", "Volando"}},
		Gif{ID: 3, URL: "https://media.tenor.com/rdevaZZ7Yd0AAAAi/eevee-evoli.gif", Likes: 8, Dislikes: 3, Tags: []string{"Bailando"}},
		Gif{ID: 4, URL: "https://media.tenor.com/IgUGgEFr_o4AAAAi/supermegaespecifictag.gif", Likes: 4, Dislikes: 1, Tags: []string{"Volando"}},
		Gif{ID: 5, URL: "https://media.tenor.com/swTLDbJLpSYAAAAi/eevee-dance.gif", Likes: 1, Dislikes: 0, Tags: []string{"Pokemon"}},
		Gif{ID: 5, URL: "https://media.tenor.com/_-Y9OYD_cWwAAAAj/duck-bwong.gif", Likes: 1, Dislikes: 0, Tags: []string{"duck", "bwong", "dance"},},		
	}

	// Guardar datos de prueba en Redis
	for _, gif := range gifs {
		serialized, _ := json.Marshal(gif)
		redisClient.Set("gif:"+strconv.Itoa(gif.ID), serialized, 0)
	}
}

var errGifNotFound = errors.New("Gif not found")

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}


// Log middleware
type LogResponseWriter struct {
    http.ResponseWriter
    statusCode int
    buf        bytes.Buffer
}

func NewLogResponseWriter(w http.ResponseWriter) *LogResponseWriter {
    return &LogResponseWriter{ResponseWriter: w}
}

func (w *LogResponseWriter) WriteHeader(code int) {
    w.statusCode = code
    w.ResponseWriter.WriteHeader(code)
}

func (w *LogResponseWriter) Write(body []byte) (int, error) {
    w.buf.Write(body)
    return w.ResponseWriter.Write(body)
}

type LogMiddleware struct {
    logger *log.Logger
}

func NewLogMiddleware(logger *log.Logger) *LogMiddleware {
    return &LogMiddleware{logger: logger}
}

func (m *LogMiddleware) Func() mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            startTime := time.Now()

            logRespWriter := NewLogResponseWriter(w)
            next.ServeHTTP(logRespWriter, r)

            m.logger.Printf(
                "duration=%s status=%d",
                time.Since(startTime).String(),
                logRespWriter.statusCode)
        })
    }
}

