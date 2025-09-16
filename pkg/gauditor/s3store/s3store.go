package s3store

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	gauditor "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Store implements gauditor.Storage by writing JSON lines to S3.
// Save appends one object per event; Query lists and filters.
// Suitable for append-only use; not optimized for massive queries.
type Store struct {
	client *s3.Client
	bucket string
	prefix string
}

// New constructs an S3-backed store.
func New(client *s3.Client, bucket, prefix string) *Store {
	prefix = strings.TrimSuffix(prefix, "/")
	return &Store{client: client, bucket: bucket, prefix: prefix}
}

func (s *Store) objectKey(e gauditor.Event) string {
	ts := e.Timestamp.UTC().Format("2006/01/02/15/04/05.000000000")
	name := fmt.Sprintf("%s-%s.json", ts, e.ID)
	if s.prefix == "" {
		return path.Join("gauditor", e.Tenant, name)
	}
	return path.Join(s.prefix, e.Tenant, name)
}

// Save uploads the event as a JSON object.
func (s *Store) Save(ctx context.Context, e gauditor.Event) (gauditor.Event, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return e, err
	}
	key := s.objectKey(e)
	uploader := manager.NewUploader(s.client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        strings.NewReader(string(body)),
		ContentType: aws.String("application/json"),
	})
	return e, err
}

// Query lists objects under tenant prefix and filters client-side.
func (s *Store) Query(ctx context.Context, q gauditor.Query) ([]gauditor.Event, error) {
	if q.Tenant == "" {
		return nil, nil
	}
	prefix := path.Join(s.prefix, q.Tenant)
	if s.prefix == "" {
		prefix = path.Join("gauditor", q.Tenant)
	}

	pager := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	var out []gauditor.Event
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			// Optional time filter using LastModified to short-circuit
			if q.Since != nil && obj.LastModified.Before(*q.Since) {
				continue
			}
			if q.Until != nil && obj.LastModified.After(*q.Until) {
				continue
			}
			// Fetch object
			get, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: obj.Key})
			if err != nil {
				return nil, err
			}
			var e gauditor.Event
			dec := json.NewDecoder(get.Body)
			_ = dec.Decode(&e)
			_ = get.Body.Close()
			// Apply precise filters
			if q.ActorID != "" && e.Actor.ID != q.ActorID {
				continue
			}
			if q.Action != "" && e.Action != q.Action {
				continue
			}
			if q.TargetID != "" && e.Target.ID != q.TargetID {
				continue
			}
			if q.Since != nil && e.Timestamp.Before((*q.Since).UTC()) {
				continue
			}
			if q.Until != nil && e.Timestamp.After((*q.Until).UTC()) {
				continue
			}
			out = append(out, e)
			if q.Limit > 0 && len(out) >= q.Limit {
				return out, nil
			}
		}
	}
	// S3 does not guarantee order; best-effort sort by timestamp
	sortByTime(out)
	return out, nil
}

func sortByTime(events []gauditor.Event) {
	if len(events) < 2 {
		return
	}
	// insertion sort for tiny slices to avoid importing sort package
	for i := 1; i < len(events); i++ {
		j := i
		for j > 0 && events[j-1].Timestamp.After(events[j].Timestamp) {
			events[j-1], events[j] = events[j], events[j-1]
			j--
		}
	}
}
