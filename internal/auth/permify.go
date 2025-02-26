// internal/auth/permify.go

package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pservice "buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
	v1 "buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
	permify_grpc "github.com/Permify/permify-go/grpc"
)

type PermifyService struct {
	client        *permify_grpc.Client
	tenant        string
	schemaVersion string
	snapToken     string
	depth         int32
}

func WitTenant(tenant string) func(*PermifyService) {
	return func(s *PermifyService) {
		s.tenant = tenant
	}
}

// WithSchemaVersion sets the schema version for the Permify service
func WithSchemaVersion(schemaVersion string) func(*PermifyService) {
	return func(s *PermifyService) {
		s.schemaVersion = schemaVersion
	}
}

// WithSnapToken sets the snap token for the Permify service
func WithSnapToken(snapToken string) func(*PermifyService) {
	return func(s *PermifyService) {
		s.snapToken = snapToken
	}
}

// WithDepth sets the depth for the Permify service
func WithDepth(depth int32) func(*PermifyService) {
	return func(s *PermifyService) {
		s.depth = depth
	}
}

// NewPermifyService creates a new Permify service
func NewPermifyService(host string, options ...func(*PermifyService)) (*PermifyService, error) {
	client, err := permify_grpc.NewClient(
		permify_grpc.Config{
			Endpoint: host,
		},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	service := &PermifyService{client: client, schemaVersion: "v1", depth: 50}
	for _, o := range options {
		o(service)
	}

	if service.tenant == "" {
		service.tenant = "t1"
	}

	return service, nil
}

type Resource struct {
	Type string
	ID   string
}

type Entity Resource
type Subject Resource

// CheckPermission checks if a subject has a permission on an entity
func (s *PermifyService) CheckPermission(entity Entity, permission string, subject Subject) (bool, error) {
	cr, err := s.client.Permission.Check(context.Background(), &v1.PermissionCheckRequest{
		TenantId: s.tenant,
		Metadata: &v1.PermissionCheckRequestMetadata{
			SnapToken:     s.snapToken,
			SchemaVersion: s.schemaVersion,
			Depth:         s.depth,
		},
		Entity: &v1.Entity{
			Type: entity.Type,
			Id:   entity.ID,
		},
		Permission: permission,
		Subject: &v1.Subject{
			Type: subject.Type,
			Id:   subject.ID,
		},
	})
	if err != nil {
		return false, err
	}

	if cr.Can == pservice.CheckResult_CHECK_RESULT_ALLOWED {
		return true, nil
	}

	return false, nil
}

func (s *PermifyService) WriteRelationship(entity Entity, relation string, subject Subject) error {
	_, err := s.client.Data.WriteRelationships(context.Background(), &v1.RelationshipWriteRequest{
		TenantId: s.tenant,
		Metadata: &v1.RelationshipWriteRequestMetadata{
			SchemaVersion: s.schemaVersion,
		},
		Tuples: []*v1.Tuple{
			{
				Entity: &v1.Entity{
					Type: entity.Type,
					Id:   entity.ID,
				},
				Relation: relation,
				Subject: &v1.Subject{
					Type: subject.Type,
					Id:   subject.ID,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *PermifyService) DeleteRelationship(entity Entity, relation string, subject Subject) error {
	_, err := s.client.Data.DeleteRelationships(context.Background(), &v1.RelationshipDeleteRequest{
		TenantId: s.tenant,
		Filter: &v1.TupleFilter{
			Entity: &v1.EntityFilter{
				Type: entity.Type,
				Ids:  []string{entity.ID},
			},
			Relation: relation,
			Subject: &v1.SubjectFilter{
				Type: subject.Type,
				Ids:  []string{subject.ID},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
