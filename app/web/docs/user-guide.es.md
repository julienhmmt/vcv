# Guía de usuario - VaultCertsViewer (VCV)

## ¿Qué es VCV?

VaultCertsViewer (VCV) es una interfaz web ligera diseñada para visualizar y monitorear los certificados gestionados por los motores PKI de HashiCorp Vault (u OpenBao). Proporciona un panel centralizado para rastrear las fechas de vencimiento, el estado (válido, vencido, revocado) y los detalles técnicos de sus certificados en múltiples instancias de Vault y puntos de montaje PKI.

## Funcionalidades

- **Soporte multi-vault**: Conéctese a una o varias instancias de Vault simultáneamente.
- **Selector de motores PKI**: Filtre certificados por instancia de Vault y punto de montaje PKI mediante un modal interactivo con búsqueda, selección/deselección por vault o globalmente.
- **Panel de control**: Gráfico de anillo con estadísticas en tiempo real sobre la distribución del estado de los certificados (válido, por vencer, vencido, revocado). Haga clic en un segmento o tarjeta de estado para filtrar la tabla instantáneamente.
- **Búsqueda y filtrado**: Busque por nombre común (CN) o nombres alternativos del sujeto (SAN). Filtre por estado mediante las tarjetas del panel de control.
- **Ordenación**: Ordene la tabla de certificados por nombre común, fecha de creación, fecha de vencimiento, nombre del Vault o punto de montaje PKI. Haga clic en un encabezado de columna para alternar entre orden ascendente/descendente.
- **Paginación**: Paginación del lado del servidor con tamaños de página configurables (25, 50, 100 o Todos).
- **Vista detallada**: Acceda a los metadatos completos del certificado en un modal organizado: identidad (sujeto, emisor, número de serie, SANs), fechas de validez con estado de expiración, detalles técnicos (algoritmo de clave, uso de clave, huellas digitales SHA-1/SHA-256) y contenido PEM para visualización.
- **Estado de Vault**: Un indicador en el encabezado (icono de escudo con punto de estado) muestra el estado de conexión en tiempo real de sus instancias de Vault. Haga clic para abrir un modal detallado con información de salud por vault y un botón de actualización.
- **Notificaciones de vencimiento**: Un banner en la parte superior de la página advierte sobre certificados que vencen dentro de los umbrales configurados (crítico / advertencia).
- **Notificaciones toast**: Mensajes toast en tiempo real para cambios de conexión de Vault, errores y retroalimentación del usuario.
- **Caché y actualización**: Los datos de certificados se almacenan en caché del lado del servidor (TTL de 15 min). Use el botón de actualización (↻) en el encabezado para invalidar la caché y obtener datos frescos.
- **Documentación integrada**: Acceda a esta guía de usuario y a la referencia de configuración directamente desde la interfaz mediante el botón de documentación (📖).
- **Sincronización de URL**: Los filtros, búsqueda, orden, paginación y selección de montajes se reflejan en la URL para compartir y guardar en favoritos.
- **I18n**: Soporte completo para inglés, francés, español, alemán e italiano. Cambie el idioma con el menú desplegable en el encabezado.
- **Modo oscuro**: Interfaz moderna con interruptor persistente de modo oscuro/claro.
- **Panel de administración**: Gestione `settings.json` visualmente (agregar/eliminar instancias Vault, configurar umbrales, registro, CORS). Requiere una contraseña de administrador configurada en `settings.json`.
- **Métricas Prometheus**: Exponga métricas de certificados y conexión en `/metrics` para monitoreo y alertas.

## Uso de la interfaz

### Panel de control

El panel de control muestra un gráfico de anillo y cuatro tarjetas de estado (Válido, Por vencer, Vencido, Revocado). Haga clic en cualquier tarjeta o segmento del gráfico para filtrar la tabla de certificados por ese estado. Aparece un botón «Borrar filtro» para restablecer el filtro.

### Selector de motores PKI

Haga clic en el botón «PKI Engines» en la barra de filtros para abrir el modal de selección de montajes. Los montajes están agrupados por instancia de Vault. Puede:

- Buscar montajes por nombre.
- Seleccionar/deseleccionar montajes individuales.
- Seleccionar/deseleccionar todos los montajes de una instancia de Vault específica.
- Seleccionar/deseleccionar todos los montajes globalmente.

La tabla de certificados se actualiza automáticamente al alternar montajes.

### Detalles del certificado

Haga clic en el botón «Detalles» en cualquier fila para abrir un modal con los metadatos completos del certificado, organizados en tres secciones: identidad (sujeto, emisor, número de serie, SANs), validez (fechas de creación/expiración con cuenta regresiva), y detalles técnicos (algoritmo de clave, uso de clave, huellas digitales SHA-1/SHA-256).

### Estado de Vault

El icono de escudo en el encabezado muestra el estado global de conexión de Vault (verde = todos conectados, rojo = al menos uno desconectado). Haga clic para ver el estado por vault. Puede forzar una verificación de salud desde el modal.

## Configuración

VCV se configura principalmente mediante un archivo `settings.json`. El panel de administración permite editar este archivo visualmente. Consulte la documentación de configuración para todos los detalles.

Todos los ajustes de la aplicación (instancias Vault, umbrales de vencimiento, registro, CORS, etc.) están definidos en `settings.json`. El panel de administración le permite gestionar estos ajustes visualmente a través de la interfaz web.

> **Nota:** El panel de administración requiere que se configure una contraseña de administrador en el archivo `settings.json` bajo el campo `admin.password`.

## Límites y lo que NO hace

- **Solo lectura**: VCV es una herramienta de visualización. **No** permite emitir, renovar o revocar certificados.
- **Autenticación**: VCV asume que ha proporcionado tokens válidos para las instancias de Vault a las que se conecta.

## Soporte

Para problemas o solicitudes de funciones, consulte el repositorio del proyecto.
