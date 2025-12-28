# Documentación del usuario - VaultCertsViewer (VCV)

## ¿Qué es VCV?

VaultCertsViewer (VCV) es una interfaz web ligera diseñada para visualizar y monitorear los certificados gestionados por los motores PKI de HashiCorp Vault. Proporciona un panel centralizado para rastrear las fechas de vencimiento, el estado (válido, vencido, revocado) y los detalles técnicos de sus certificados en múltiples instancias de Vault y puntos de montaje PKI.

## Capacidades

- **Soporte multi-vault**: Conéctese a una o varias instancias de Vault.
- **Descubrimiento de motores PKI**: Descubre automáticamente los puntos de montaje PKI a los que tiene acceso.
- **Panel de control**: Estadísticas en tiempo real sobre la distribución del estado de los certificados y el cronograma de vencimiento.
- **Búsqueda y filtrado**: Busque por nombre común (CN) o nombres alternativos del sujeto (SAN). Filtre por Vault, punto de montaje PKI, estado o umbral de vencimiento.
- **Vista detallada**: Acceda a los metadatos completos del certificado, incluidos el emisor, las huellas digitales y el contenido PEM.
- **Exportar**: Descargue archivos PEM de certificados directamente desde la interfaz de usuario.
- **I18n**: Soporte completo para inglés, francés, español, alemán e italiano.
- **Modo oscuro**: Interfaz de usuario moderna con interruptor de modo oscuro/claro.

## Configuración

VCV se configura principalmente a través de variables de entorno o un archivo `settings.json`.

### Principales variables de entorno

- `VAULT_ADDRS`: Lista de direcciones de Vault separadas por comas.
- `VCV_EXPIRE_WARNING`: Umbral en días para notificaciones de advertencia (predeterminado: 30).
- `VCV_EXPIRE_CRITICAL`: Umbral en días para notificaciones críticas (predeterminado: 7).
- `LOG_LEVEL`: Detalle del registro (info, debug, error).

## Límites y lo que NO hace

- **Solo lectura**: VCV es actualmente una herramienta de visualización. **No** permite emitir, renovar o revocar certificados.
- **Autenticación**: VCV asume que ha proporcionado tokens válidos o ha configurado la autenticación para las instancias de Vault a las que se conecta.
- **Gestión de Vault**: No gestiona las políticas de Vault ni la configuración de PKI; solo lee los datos existentes.

## Soporte

Para problemas o solicitudes de funciones, consulte el repositorio del proyecto.
