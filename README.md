# mongo_nats_worker

Este repositorio implementa un worker en Go que recibe mensajes de NATS y los almacena en MongoDB. Es útil para procesar y almacenar documentos con coordenadas geoespaciales enviados a través de NATS.

## Funcionamiento principal

- El worker se conecta a un servidor NATS (por defecto `nats://localhost:4222`).
- Se suscribe al tópico `coordinates`.
- Cada mensaje recibido se interpreta como un documento JSON con estructura geoespacial y metadatos (ver abajo).
- El documento se inserta en una colección MongoDB (por defecto, en la base de datos `test`, colección `coordinates`).

**Estructura de documento:**
```go
type Document struct {
	UniqueId     string
	UserId       string
	Fleet        string
	Location     MongoLocation // incluye tipo y coordenadas
	OriginIp     string
	LastModified int64
}
type MongoLocation struct {
	Type        string
	Coordinates []float64
}
```

## Modo de uso

1. **Configura tus variables de entorno** (opcional, se usan valores por defecto si no hay `.env`):
   - `NATS_URL` (por defecto: `nats://localhost:4222`)
   - `MONGO_URI` (por defecto: `mongodb://localhost:27017`)
   - `DATABASE_NAME` (por defecto: `test`)
   - `COORDINATE_COLLECTION_NAME` (por defecto: `coordinates`)
   - `HOOK_COLLECTION_NAME` (por defecto: `hooks`)

2. **Compila y ejecuta manualmente:**
   ```bash
   go build -o mongo_nats_worker
   ./mongo_nats_worker
   ```

3. **Usa Docker:**
   ```bash
   docker build --pull --rm -f Dockerfile -t mongo_nats_worker:latest .
   docker run --env-file .env mongo_nats_worker:latest
   ```

4. **Ejemplo de uso con Docker Compose:**
   ```yaml
   version: "3.8"
   services:
     nats:
       image: nats:latest
       ports:
         - "4222:4222"
     mongo:
       image: mongo:latest
       ports:
         - "27017:27017"
     worker:
       image: mongo_nats_worker:latest
       depends_on:
         - nats
         - mongo
       environment:
         NATS_URL: "nats://nats:4222"
         MONGO_URI: "mongodb://mongo:27017"
         DATABASE_NAME: "test"
         COORDINATE_COLLECTION_NAME: "coordinates"
   ```

## Observaciones

- El worker permanece activo indefinidamente esperando mensajes.
- Si faltan variables de entorno, se usan valores por defecto, pero se recomienda configurarlas.
- El contenedor es minimalista (`FROM scratch`) y sólo ejecuta el binario compilado.
- Se espera que los mensajes publicados en NATS bajo el tópico `coordinates` tengan el formato JSON del documento mostrado arriba, incluyendo el campo `location` con tipo y coordenadas.

---

**Repositorio:** [120m4n/mongo_nats_worker](https://github.com/120m4n/mongo_nats_worker)
