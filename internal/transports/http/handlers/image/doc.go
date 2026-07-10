// Package image implements the product-image upload/serve endpoints. This is
// plain HTTP, not gRPC — see the CRM TD: images have no dedicated entity,
// gRPC method, or metadata store; a file is uploaded, saved on disk under a
// generated id, and that id is attached to Product.imageIds via the
// ProductsService Update RPC by the caller.
package image
