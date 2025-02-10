### AutoRest Configuration

> see <https://aka.ms/autorest>

```yaml
go: true
track2: true
input-file: msi-credentials-data-plane.openapi.v2.json
output-folder: client
use-extension:
  "@autorest/modelerfour": "~4.27.0"
  "@autorest/go": "4.0.0-preview.69"
```