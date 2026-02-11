# Referencia de configuración

## 📋 Descripción general

VaultCertsViewer (VCV) se configura principalmente mediante un archivo `settings.json`. El panel de administración le permite gestionar este archivo directamente desde la interfaz web. Las variables de entorno son compatibles como alternativa legacy cuando no se encuentra un archivo `settings.json`.

VCV utiliza una arquitectura de renderizado del lado del servidor basada en [HTMX](https://htmx.org/). Todos los filtrados, ordenaciones y paginaciones se gestionan del lado del servidor para un rendimiento óptimo.

> **⚠️ Importante:** Después de guardar los cambios, puede ser necesario reiniciar el servidor para que todas las configuraciones surtan efecto.

## 🔐 Acceso al panel de administración

### Contraseña de admin en el archivo de configuración

Para activar el panel de administración, añada una sección `admin` a su archivo `settings.json` con una contraseña bcrypt:

```json
{
  "admin": {
    "password": "$2b$12$su-hash-bcrypt-aqui"
  }
}
```

La contraseña debe ser un **hash bcrypt** (prefijo `$2a$`, `$2b$`, o `$2y$`).

```bash
# Generar un hash bcrypt (ejemplo con htpasswd)
htpasswd -nbBC 10 admin SuContraseñaSegura | cut -d: -f2

# O con Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'SuContraseña', bcrypt.gensalt()).decode())"
```

También puede utilizar el servicio 'bcrypt' de <https://tools.hommet.net/bcrypt> para generar un hash bcrypt (no se almacenan datos).

**Nombre de usuario predeterminado:** `admin` (no editable, valor predeterminado)
**Duración de la sesión:** 12 horas (no editable, valor predeterminado)
**Limitación de intentos de inicio de sesión:** 10 intentos por 5 minutos (no editable, valor predeterminado)

## 📁 Configuración de la aplicación

### Entorno (app.env)

Define el entorno de la aplicación. Afecta las funciones de seguridad y el comportamiento del registro.

- `dev` - Modo desarrollo (registro detallado, sin limitación de velocidad)
- `prod` - Modo producción (cookies seguras, limitación de velocidad activada)

**Predeterminado:** `prod`

### Puerto (app.port)

Puerto de escucha del servidor HTTP.

**Predeterminado:** `52000`

### Ruta del archivo de configuración

VCV busca el archivo de configuración en este orden:

1. `settings.dev.json`
2. `settings.prod.json`
3. `settings.json`
4. `./settings.json`
5. `/app/settings.json`

### Registro (app.logging)

Configure el comportamiento del registro de la aplicación:

- **level**: `debug`, `info`, `warn`, `error`
- **format**: `json` o `text`
- **output**: `stdout`, `file` o `both`
- **file_path**: Ruta del archivo de registro cuando output es `file` o `both`

**Predeterminados:**

- level: `info`
- format: `json`
- output: `stdout`
- file_path: `/var/log/app/vcv.log`

## 📜 Configuración de certificados

### Umbrales de expiración (certificates.expiration_thresholds)

Configure cuándo los certificados se marcan como próximos a expirar:

- **critical**: Días antes de la expiración para mostrar alerta crítica
- **warning**: Días antes de la expiración para mostrar advertencia

Estos umbrales controlan:

- Banner de notificación en la parte superior de la página
- Codificación de colores en la tabla de certificados (rojo para crítico, amarillo para advertencia)
- Visualización de la línea de tiempo en el panel de control
- Métricas Prometheus (`vcv_certificates_expiring_critical`, `vcv_certificates_expiring_warning`)

**Predeterminados:**

- critical: `7`
- warning: `30`

## 🌐 Configuración CORS (cors)

### Orígenes permitidos (cors.allowed_origins)

Array de orígenes CORS permitidos. Use `["*"]` para permitir todos los orígenes (no recomendado en producción).

**Ejemplo:**

```json
"allowed_origins": ["https://example.com", "https://app.example.com"]
```

### Permitir credenciales (cors.allow_credentials)

Booleano para permitir credenciales en solicitudes CORS.

**Predeterminado:** `false`

**Nota:** CORS es principalmente útil si integra VCV en otra aplicación web o accede desde un dominio diferente.

## 🔐 Configuración de Vault

### Múltiples instancias de Vault

VaultCertsViewer soporta la monitorización de múltiples instancias de Vault simultáneamente. Cada instancia de Vault requiere:

- **ID**: Identificador único para esta instancia de Vault (requerido)
- **Display name**: Nombre legible mostrado en la interfaz (opcional)
- **Address**: URL del servidor Vault (ej: `https://vault.example.com:8200`)
- **Token**: Token de Vault de solo lectura con acceso PKI (requerido)
- **PKI mounts**: Array de rutas de montaje PKI (ej: `["pki", "pki2", "pki-prod"]`)
- **Enabled**: Si esta instancia de Vault está activa

### Configuración TLS

Para Vaults que utilizan certificados CA personalizados o autofirmados:

- **TLS CA cert (Base64)**: Bundle CA PEM codificado en base64 (método preferido)
- **TLS CA cert path**: Ruta del archivo al bundle CA PEM
- **TLS CA path**: Directorio que contiene certificados CA
- **TLS server name**: Anulación del nombre de servidor SNI
- **TLS insecure**: Omitir verificación TLS (⚠️ solo desarrollo, no recomendado)

```bash
# Codificar un certificado en base64
cat ruta-al-cert.pem | base64 | tr -d '\n'
```

**Precedencia:** Si `tls_ca_cert_base64` está configurado, tiene prioridad sobre las rutas de archivo.

### Permisos del token de Vault

El token de Vault debe tener acceso de solo lectura a los montajes PKI. Ejemplo de política:

```hcl
# Montajes explícitos (recomendado para producción)
vault policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "pki2/certs"   { capabilities = ["list"] }
path "pki2/certs/*" { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF

# Patrón con comodín (para entornos dinámicos)
vault policy write vcv - <<'EOF'
path "pki*/certs"    { capabilities = ["list"] }
path "pki*/certs/*"  { capabilities = ["read","list"] }
path "sys/health"    { capabilities = ["read"] }
EOF

# Crear token
vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

Debe reemplazar 'pki' y 'pki2' con las rutas de montaje PKI de su Vault. Agregue tantas rutas de montaje PKI como tenga en su Vault.

## ⚡ Optimizaciones de rendimiento

### Caché

VaultCertsViewer implementa caché para mejorar el rendimiento:

- **TTL de caché de certificados:** 15 minutos (reduce las llamadas a la API de Vault)
- **Caché de verificación de salud:** 30 segundos (para el indicador de estado en el encabezado)
- **Obtención paralela:** Múltiples Vaults se consultan simultáneamente
- **Invalidación de caché:** Use el botón de actualización (↻) en el encabezado o `POST /api/cache/invalidate` para vaciar la caché de certificados

Con múltiples Vaults, la obtención paralela proporciona tiempos de carga **3-10× más rápidos** comparados con consultas secuenciales.

## 📊 Monitorización y métricas

### Métricas Prometheus

Disponibles en el endpoint `/metrics`:

- `vcv_certificates_total` - Número total de certificados
- `vcv_certificates_valid` - Número de certificados válidos
- `vcv_certificates_expired` - Número de certificados vencidos
- `vcv_certificates_revoked` - Número de certificados revocados
- `vcv_certificates_expiring_critical` - Certificados que vencen dentro del umbral crítico
- `vcv_certificates_expiring_warning` - Certificados que vencen dentro del umbral de advertencia
- `vcv_vault_connected` - Estado de conexión de Vault (1=conectado, 0=desconectado)
- `vcv_cache_size` - Número de entradas en caché
- `vcv_last_fetch_timestamp` - Marca de tiempo Unix de la última obtención de certificados

Todas las métricas incluyen etiquetas: `vault_id`, `vault_name`, `pki_mount`

### Endpoints de salud y API

- `/api/health` - Verificación de salud básica (siempre devuelve 200 OK)
- `/api/ready` - Sonda de disponibilidad (verifica el estado de la aplicación)
- `/api/status` - Estado detallado incluyendo todas las conexiones de Vault
- `/api/version` - Información de versión de la aplicación
- `/api/config` - Configuración de la aplicación (umbrales de expiración, lista de vaults)
- `/api/i18n` - Traducciones para el idioma actual
- `/api/certs` - Lista de certificados (JSON)
- `/api/certs/{id}/details` - Detalles del certificado (JSON)
- `/api/certs/{id}/pem` - Contenido PEM del certificado (JSON)
- `/api/certs/{id}/pem/download` - Descargar archivo PEM del certificado
- `POST /api/cache/invalidate` - Invalidar la caché de certificados

### Limitación de velocidad

En modo `prod`, la limitación de velocidad de la API está activada a **300 solicitudes por minuto** por cliente. Los siguientes caminos están exentos:

- `/api/health`, `/api/ready`, `/metrics`
- `/assets/*` (archivos estáticos)

## 🔒 Mejores prácticas de seguridad

- Utilice siempre el entorno `prod` en producción
- Proteja el archivo `settings.json` (contiene tokens sensibles)
- Use tokens de Vault de solo lectura con permisos mínimos
- La limitación de velocidad es automática en modo `prod` (300 sol./min)
- La protección CSRF está activada en todas las solicitudes que modifican el estado
- Los encabezados de seguridad (X-Content-Type-Options, X-Frame-Options, etc.) se establecen automáticamente
- Ejecute el contenedor con `--read-only` y `--cap-drop=ALL`

## 📝 Ejemplo settings.json

```json
{
  "app": {
    "env": "prod",
    "port": 52000,
    "logging": {
      "level": "info",
      "format": "json",
      "output": "stdout",
      "file_path": "/var/log/app/vcv.log"
    }
  },
  "certificates": {
    "expiration_thresholds": {
      "critical": 7,
      "warning": 30
    }
  },
  "cors": {
    "allowed_origins": ["https://example.com"],
    "allow_credentials": false
  },
  "vaults": [
    {
      "id": "vault-prod",
      "display_name": "Vault Producción",
      "address": "https://vault.example.com:8200",
      "token": "hvs.xxx",
      "pki_mounts": ["pki", "pki-intermediate"],
      "enabled": true,
      "tls_insecure": false,
      "tls_ca_cert_base64": "LS0tLS1CRUdJTi...",
      "tls_server_name": "vault.example.com"
    },
    {
      "id": "vault-dev",
      "display_name": "Vault Desarrollo",
      "address": "https://vault-dev.example.com:8200",
      "token": "hvs.yyy",
      "pki_mounts": ["pki_dev"],
      "enabled": true,
      "tls_insecure": true
    }
  ]
}
```

> **💡 Consejo:** Use el panel de administración para editar estas configuraciones visualmente. Los cambios se guardan en el archivo `settings.json`.
