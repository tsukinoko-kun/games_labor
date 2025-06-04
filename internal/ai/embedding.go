package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/genai"
	"slices"
)

var qdrantClient *qdrant.Client

const (
	embeddingModel       = "text-embedding-004"
	qdrantCollectionName = "gameslabor"
	dimensions           = 768
)

func init() {
	var err error
	qdrantClient, err = qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to Qdrant: %v\n", err)
		os.Exit(1)
		return
	}

	ctx := context.Background()

	collectionsResponse, err := qdrantClient.ListCollections(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list collections: %v\n", err)
		os.Exit(1)
		return
	}

	collectionExists := slices.Contains(collectionsResponse, qdrantCollectionName)

	if !collectionExists {
		err = qdrantClient.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: qdrantCollectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     dimensions,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create collection: %v\n", err)
			os.Exit(1)
			return
		}
	}
}

func (ai *AI) Embed(contents []*genai.Content) error {
	resp, err := ai.llmClient.Models.EmbedContent(ai.ctx, embeddingModel, contents, &genai.EmbedContentConfig{
		TaskType:             "RETRIEVAL_QUERY",
		OutputDimensionality: genai.Ptr[int32](dimensions),
	})
	if err != nil {
		return err
	}

	var pointsToUpsert []*qdrant.PointStruct
	for _, embedding := range resp.Embeddings {
		pointsToUpsert = append(pointsToUpsert, &qdrant.PointStruct{
			Id:      nextQdrantId(),
			Vectors: qdrant.NewVectors(embedding.Values...),
		})
	}

	_, err = qdrantClient.Upsert(
		ai.ctx,
		&qdrant.UpsertPoints{
			CollectionName: qdrantCollectionName,
			Points:         pointsToUpsert,
			Wait:           qdrant.PtrOf(true),
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func nextQdrantId() *qdrant.PointId {
	return &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: uuid.NewString()}}
}
