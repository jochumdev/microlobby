package utils

import (
	"context"
	"net/http"
	"strings"

	"go-micro.dev/v4/metadata"
)

// RequestToContext adds HTTP request headers to the context as metadata
func RequestToContext(ctx context.Context, r *http.Request) context.Context {
	md := make(metadata.Metadata, len(r.Header))
	for k, v := range r.Header {
		if k == "Authorization" {
			k = "X-Microlobby-Authorization"
		}
		md[k] = strings.Join(v, ",")
	}
	return metadata.NewContext(ctx, md)
}
