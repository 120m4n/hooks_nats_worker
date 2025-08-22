package main

import (
	"context"
	"encoding/json"
	"log"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/120m4n/mongo_nats/config"
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
	log.Printf("Worker iniciado - DB: %s, Collection: %s, Workers: 5", cfg.DatabaseName, cfg.HookCollectionName)

	// Connect to NATS server
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Get a handle for your collection usando configuración
	collection := mongoClient.Database(cfg.DatabaseName).Collection(cfg.HookCollectionName)

	// Canal para documentos recibidos
	// ...eliminado canal docsChan obsoleto...

	// Pool de workers
	numWorkers := 5                                    // puedes ajustar este valor
	hookChan := make(chan map[string]interface{}, 100) // buffer configurable
	startHookWorkerPool(numWorkers, hookChan, collection)

	// Iniciar reporte de estadísticas cada 120 segundos
	go startStatsReporter()

	// Subscribe to "hooks" topic
	if err := subscribeHooks(nc, hookChan); err != nil {
		log.Fatalf("Error subscribing to topic: %v", err)
	}

	// Mantener el proceso vivo
	select {}
}

// startHookWorkerPool launches a pool of workers to process generic hook events
func startHookWorkerPool(numWorkers int, hookChan <-chan map[string]interface{}, collection *mongo.Collection) {
	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			for event := range hookChan {
				processHookEvent(id, event, collection)
			}
		}(i)
	}
}

// processHookEvent handles the insertion of a generic hook event into MongoDB
func processHookEvent(id int, event map[string]interface{}, collection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, event)
	if err != nil {
		atomic.AddInt64(&errorCount, 1)
		log.Printf("Worker %d: Error inserting hook event into MongoDB: %v", id, err)
	} else {
		atomic.AddInt64(&processedCount, 1)
	}
}

// startStatsReporter reporta estadísticas cada 30 segundos
func startStatsReporter() {
	ticker := time.NewTicker(120 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		processed := atomic.LoadInt64(&processedCount)
		errors := atomic.LoadInt64(&errorCount)
		validationErrs := atomic.LoadInt64(&validationErrors)
		log.Printf("Stats - Procesados: %d, Errores DB: %d, Errores validación: %d",
			processed, errors, validationErrs)

		// Reset de contadores para evitar crecimiento indefinido
		atomic.StoreInt64(&processedCount, 0)
		atomic.StoreInt64(&errorCount, 0)
		atomic.StoreInt64(&validationErrors, 0)
	}
}

// subscribeHooks subscribes to the "hooks" topic and sends valid events to hookChan
func subscribeHooks(nc *nats.Conn, hookChan chan<- map[string]interface{}) error {
	_, err := nc.Subscribe("hooks", func(m *nats.Msg) {
		var event map[string]interface{}
		if err := json.Unmarshal(m.Data, &event); err != nil {
			log.Printf("Error unmarshalling hook event: %v", err)
			return
		}
		// Validación mínima: debe tener metadata
		if _, ok := event["metadata"]; !ok {
			atomic.AddInt64(&validationErrors, 1)
			log.Printf("Evento sin metadata, descartado")
			return
		}
		// Enviar el evento al canal para procesamiento concurrente
		hookChan <- event
	})
	return err
}
