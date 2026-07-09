// Package commands defines the root Cobra CLI tree for the my-business-crm binary.
//
// The tree is assembled in [New] and currently looks like:
//
//	my-business-crm       (root)
//	├── server            server lifecycle
//	│   └── run           start the gRPC + HTTP server
//	└── version           print build / version info
//
// Call [Run] to build the tree, set CLI args, and execute.
package commands
