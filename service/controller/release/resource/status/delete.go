package status

import "context"

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	// Release.Status.Cycle is not important when the release is deleted.
	// Also Release CRs are never deleted in production environments.
	return nil
}
