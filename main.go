package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
	// "go.mongodb.org/mongo-driver/bson"
	"errors"

	"github.com/120m4n/mongo_nats/config"
	"github.com/120m4n/mongo_nats/model"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Contadores globales para estadísticas
var (
	processedCount   int64
	errorCount       int64
	validationErrors int64
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Log inicial de configuración (solo al iniciar)
	log.Printf("Worker iniciado - DB: %s, Collection: %s, Workers: 5", cfg.DatabaseName, cfg.Coor_CollectionName)

	// Connect to NATS server
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Get a handle for your collection usando configuración
	collection := mongoClient.Database(cfg.DatabaseName).Collection(cfg.Coor_CollectionName)

	// Canal para documentos recibidos
	docsChan := make(chan model.Document, 100) // buffer configurable

	// Pool de workers
	numWorkers := 5 // puedes ajustar este valor
	startWorkerPool(numWorkers, docsChan, collection)

	// Iniciar reporte de estadísticas cada 30 segundos
	go startStatsReporter()

	// Subscribe to "coordinates" topic
	if err := subscribeCoordinates(nc, docsChan); err != nil {
		log.Fatalf("Error subscribing to topic: %v", err)
	}

	// Mantener el proceso vivo
	select {}
}

// startWorkerPool launches a pool of workers to process documents
func startWorkerPool(numWorkers int, docsChan <-chan model.Document, collection *mongo.Collection) {
	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			for doc := range docsChan {
				processDocument(id, doc, collection)
			}
		}(i)
	}
}

// processDocument handles the insertion of a document into MongoDB
func processDocument(id int, doc model.Document, collection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		atomic.AddInt64(&errorCount, 1)
		log.Printf("Worker %d: Error inserting into MongoDB: %v", id, err)
	} else {
		atomic.AddInt64(&processedCount, 1)
	}
	// Removido el log de éxito para mejorar rendimiento
}

// startStatsReporter reporta estadísticas cada 30 segundos
func startStatsReporter() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		processed := atomic.LoadInt64(&processedCount)
		errors := atomic.LoadInt64(&errorCount)
		validationErrs := atomic.LoadInt64(&validationErrors)
		
		log.Printf("Stats - Procesados: %d, Errores DB: %d, Errores validación: %d", 
			processed, errors, validationErrs)
	}
}

// subscribeCoordinates subscribes to the "coordinates" topic and sends valid documents to docsChan
func subscribeCoordinates(nc *nats.Conn, docsChan chan<- model.Document) error {
	_, err := nc.Subscribe("coordinates", func(m *nats.Msg) {
		var doc model.Document
		if err := json.Unmarshal(m.Data, &doc); err != nil {
			log.Printf("Error unmarshalling data: %v", err)
			return
		}
		// Validar el documento antes de enviarlo al canal
		if err := validateDocument(doc); err != nil {
			atomic.AddInt64(&validationErrors, 1)
			// Solo loguear errores de validación críticos ocasionalmente
			return
		}
		// Enviar el documento al canal para procesamiento concurrente
		docsChan <- doc
	})
	return err
}

// validateDocument verifica los campos obligatorios y formato básico
func validateDocument(doc model.Document) error {
	if doc.UniqueId == "" {
		return errors.New("UniqueId vacío")
	}
	if doc.UserId == "" {
		return errors.New("UserId vacío")
	}
	if doc.Fleet == "" {
		return errors.New("Fleet vacío")
	}
	if doc.Location.Type == "" {
		return errors.New("Location.Type vacío")
	}
	if len(doc.Location.Coordinates) != 2 {
		return errors.New("Location.Coordinates debe tener longitud 2 (lat,lon)")
	}
	// Puedes agregar más validaciones según tu modelo
	return nil
}
