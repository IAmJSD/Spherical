package db

import "context"

// InternallyDeleteToken is used to internally delete a token. This is used by the token delete job.
// Returns the deletion job ID and any errors.
func InternallyDeleteToken(ctx context.Context, token string) (string, error) {
	query := "DELETE FROM sessions WHERE token = $1 RETURNING deletion_job_id"
	var deletionJobID string
	err := dbConn().QueryRow(ctx, query, token).Scan(&deletionJobID)
	return deletionJobID, err
}

// InternallyGetTokenJobID is used to get the job ID of a token. This is used by the token extend job.
func InternallyGetTokenJobID(ctx context.Context, token string) (string, error) {
	query := "SELECT deletion_job_id FROM sessions WHERE token = $1"
	var deletionJobID string
	err := dbConn().QueryRow(ctx, query, token).Scan(&deletionJobID)
	return deletionJobID, err
}
