// Package mongo implements productattributedefinitions.Storage against
// MongoDB. Read-only from the app's point of view — the seed migration in
// this package is the only writer, added to by a developer via a follow-up
// migration when a new characteristic is needed.
package mongo
