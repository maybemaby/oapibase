package gen

import "net/http"

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
      url: '/docs/swagger.json',
      dom_id: '#swagger-ui',
    });
  };
</script>
</body>
</html>
`

func RenderSwaggerUI(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func HandleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	RenderSwaggerUI(w)
}

func HandleSwaggerJson(w http.ResponseWriter, r *http.Request) {
	swagger, err := GetSwagger()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	spec, err := swagger.MarshalJSON()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(spec)
}
