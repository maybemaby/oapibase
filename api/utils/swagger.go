package utils

import (
	"fmt"
	"net/http"
)

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="SwaggerUI" />
  <title>SwaggerUI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
<script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: '%s',
      dom_id: '#swagger-ui',
    });
  };
</script>
</body>
</html>
`

// RenderSwaggerUI writes the Swagger UI html to the response writer
// swaggerPath should point to an endpoint with a valid swagger configuration
func RenderSwaggerUI(w http.ResponseWriter, swaggerPath string) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, html, swaggerPath)
}

